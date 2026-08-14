package loggingsplunk

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// authenticationRequired enforces that a token is set, directly or via
// FASTLY_SPLUNK_TOKEN, matching the live provider's Required+EnvDefaultFunc
// behavior — the schema Default alone only supplies a fallback value, not a
// requirement.
type authenticationRequired struct{}

func (authenticationRequired) Description(_ context.Context) string {
	return "authentication must set token, directly or via FASTLY_SPLUNK_TOKEN"
}

func (v authenticationRequired) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (authenticationRequired) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsUnknown() {
		return
	}

	if effectiveAuthToken(ctx, req.ConfigValue) != "" {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Missing Splunk authentication credentials",
		"`authentication` must set `token` — directly or via the FASTLY_SPLUNK_TOKEN environment variable.",
	)
}

// effectiveAuthToken returns the configured token, falling back to
// FASTLY_SPLUNK_TOKEN when omitted (an explicit "" is not rescued by the
// env var).
func effectiveAuthToken(ctx context.Context, obj types.Object) string {
	if !obj.IsNull() && !obj.IsUnknown() {
		if v, ok := obj.Attributes()["token"]; ok {
			if sv, ok := v.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
				return sv.ValueString()
			}
		}
	}
	return envStringDefault(ctx, splunkTokenEnvVar).ValueString()
}

// ValidateNoVCLOnlyAttributesForCompute returns an error diagnostic if format,
// format_version, placement, or response_condition are explicitly configured on
// a Compute service. The standalone fastly_service_logging_splunk resource
// has one schema shared by both service types — unlike the nested blocks,
// which have distinct VCL (NestedBlockSchema) and Compute
// (ComputeNestedBlockSchema) schemas — so this is the only way to catch the
// mistake before it silently sends unsupported VCL-only attributes to a
// Compute service.
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
