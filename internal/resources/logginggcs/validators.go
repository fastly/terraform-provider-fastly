package logginggcs

import (
	"context"
	"strings"

	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// gzipLevelCodecConflict enforces that gzip_level and compression_codec are not
// configured together. The Fastly API rejects a request that sets both, and the
// codec implies its own level (gzip defaults to 3), so the two are alternative
// ways to request compression. A blank compression_codec is not a conflict.
type gzipLevelCodecConflict struct{}

func (gzipLevelCodecConflict) Description(_ context.Context) string {
	return "gzip_level cannot be set when compression_codec is set"
}

func (v gzipLevelCodecConflict) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (gzipLevelCodecConflict) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// Only relevant when the user explicitly set gzip_level. Config values are
	// null when unconfigured (the -1 default is applied later, at plan time).
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var codec types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("compression_codec"), &codec)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A blank codec is allowed alongside gzip_level.
	if codec.IsNull() || codec.IsUnknown() || service.StringValue(codec) == "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Conflicting `gzip_level` and `compression_codec`",
		"`gzip_level` and `compression_codec` cannot be set together — the Fastly API rejects a request that specifies both.\n\n"+
			"- To compress at a specific gzip level, leave `compression_codec` unset and set `gzip_level`.\n"+
			"- To use a codec (`zstd`, `snappy`, or `gzip`), remove `gzip_level`. With `compression_codec = \"gzip\"`, the level defaults to `3`.",
	)
}

// notTrimmed rejects a string with leading or trailing whitespace (e.g.
// \n\t\r\f). The Fastly API silently mishandles a BigQuery secret_key with
// such whitespace (per loggingbigquery's notTrimmed) — GCS's secret_key is the
// same PEM private-key material, so the same check is applied here.
type notTrimmed struct{}

func (notTrimmed) Description(_ context.Context) string {
	return "value must not contain leading or trailing whitespace"
}

func (v notTrimmed) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (notTrimmed) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v := req.ConfigValue.ValueString()
	if v != strings.TrimSpace(v) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			"must not contain leading or trailing whitespace characters (e.g., \\n\\t\\r\\f). Consider trimming the value.",
		)
	}
}

// authenticationEitherOr enforces that the authentication block resolves to
// either account_name, or both email and secret_key. The Fastly API accepts
// account_name as an alternative to the email/secret_key pair, but rejects a
// request providing neither or only one of email/secret_key, so this is
// caught at plan/validate time instead. Each field's effective value falls
// back to its environment variable default (mirroring the schema Default
// handlers) so a config that legitimately relies on the environment for one
// or more fields is not flagged.
type authenticationEitherOr struct{}

func (authenticationEitherOr) Description(_ context.Context) string {
	return "authentication must set account_name, or both email and secret_key"
}

func (v authenticationEitherOr) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (authenticationEitherOr) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	if effectiveAccountName(ctx, req.ConfigValue) != "" {
		return
	}

	email := effectiveAuthValue(ctx, req.ConfigValue, "email", "FASTLY_GCS_EMAIL")
	secretKey := effectiveAuthValue(ctx, req.ConfigValue, "secret_key", "FASTLY_GCS_SECRET_KEY")
	if email != "" && secretKey != "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Missing GCS authentication credentials",
		"`authentication` must set `account_name`, or both `email` and `secret_key` — directly or via the "+
			"FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME (or FASTLY_GCS_ACCOUNT_NAME), FASTLY_GCS_EMAIL, and FASTLY_GCS_SECRET_KEY environment variables.",
	)
}

// effectiveAuthValue returns the configured value of attrName within the
// authentication object, or its environment variable default when the
// config leaves that field unset. obj is req.ConfigValue — the raw,
// pre-Default config — so IsNull()/IsUnknown() alone correctly distinguish
// "omitted" (falls through to the env var) from "present, even if blank"
// (returned verbatim); a field explicitly set to "" is not the same as an
// unconfigured one, since Terraform's schema Default only fills attributes
// that are truly null, not ones explicitly set to a zero value.
func effectiveAuthValue(ctx context.Context, obj types.Object, attrName, envVar string) string {
	if !obj.IsNull() && !obj.IsUnknown() {
		if v, ok := obj.Attributes()[attrName]; ok {
			if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				return sv.ValueString()
			}
		}
	}
	return envStringDefault(ctx, envVar).ValueString()
}

// effectiveAccountName returns the configured account_name within the
// authentication object, or its environment variable default (preferring
// FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME, falling back to FASTLY_GCS_ACCOUNT_NAME)
// when the config leaves it unset. See effectiveAuthValue for why
// IsNull()/IsUnknown() alone — not also checking for a blank string — is the
// correct test here.
func effectiveAccountName(_ context.Context, obj types.Object) string {
	if !obj.IsNull() && !obj.IsUnknown() {
		if v, ok := obj.Attributes()["account_name"]; ok {
			if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				return sv.ValueString()
			}
		}
	}
	return accountNameEnvValue()
}

// ValidateNoVCLOnlyAttributesForCompute returns an error diagnostic if format,
// format_version, placement, or response_condition are explicitly configured on
// a Compute service. The standalone fastly_service_logging_gcs resource has
// one schema shared by both service types — unlike the nested blocks, which
// have distinct VCL (NestedBlockSchema) and Compute (ComputeNestedBlockSchema)
// schemas — so this is the only way to catch the mistake before it silently
// sends unsupported VCL-only attributes to a Compute service.
func ValidateNoVCLOnlyAttributesForCompute(ctx context.Context, cfg tfsdk.Config) diag.Diagnostics {
	var diags diag.Diagnostics

	var format, placement, responseCondition types.String
	var formatVersion types.Int64

	diags.Append(cfg.GetAttribute(ctx, path.Root("format"), &format)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("format_version"), &formatVersion)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("placement"), &placement)...)
	diags.Append(cfg.GetAttribute(ctx, path.Root("response_condition"), &responseCondition)...)
	if diags.HasError() {
		return diags
	}

	var configured []string
	if !format.IsNull() {
		configured = append(configured, "format")
	}
	if !formatVersion.IsNull() {
		configured = append(configured, "format_version")
	}
	if !placement.IsNull() {
		configured = append(configured, "placement")
	}
	if !responseCondition.IsNull() {
		configured = append(configured, "response_condition")
	}

	if len(configured) > 0 {
		diags.AddError(
			"VCL-only attributes not supported on Compute services",
			"The following attributes only affect generated VCL and are not supported when `service_id` refers to a Compute service: "+
				strings.Join(configured, ", ")+". Remove them from this configuration.",
		)
	}

	return diags
}
