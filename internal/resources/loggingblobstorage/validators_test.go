package loggingblobstorage

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationRequired(t *testing.T) {
	tests := []struct {
		name        string
		accountName types.String
		sasToken    types.String
		envAccount  string
		envSAS      string
		wantError   bool
	}{
		{
			name:        "both configured",
			accountName: types.StringValue("myaccount"),
			sasToken:    types.StringValue("sv=token"),
			wantError:   false,
		},
		{
			name:        "account_name only",
			accountName: types.StringValue("myaccount"),
			sasToken:    types.StringNull(),
			wantError:   true,
		},
		{
			name:        "sas_token only",
			accountName: types.StringNull(),
			sasToken:    types.StringValue("sv=token"),
			wantError:   true,
		},
		{
			name:        "neither configured, no env vars",
			accountName: types.StringNull(),
			sasToken:    types.StringNull(),
			wantError:   true,
		},
		{
			name:        "neither configured, both env vars set",
			accountName: types.StringNull(),
			sasToken:    types.StringNull(),
			envAccount:  "myaccount",
			envSAS:      "sv=token",
			wantError:   false,
		},
		{
			name:        "account_name configured, sas_token falls back to env var",
			accountName: types.StringValue("myaccount"),
			sasToken:    types.StringNull(),
			envSAS:      "sv=token",
			wantError:   false,
		},
		{
			name:        "account_name configured, sas_token unset and no env var",
			accountName: types.StringValue("myaccount"),
			sasToken:    types.StringNull(),
			wantError:   true,
		},
		{
			name:        "account_name explicitly blank is not rescued by env var",
			accountName: types.StringValue(""),
			sasToken:    types.StringValue("sv=token"),
			envAccount:  "myaccount",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FASTLY_AZURE_ACCOUNT_NAME", tt.envAccount)
			t.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", tt.envSAS)

			req := validator.ObjectRequest{
				Path:        path.Root("authentication"),
				ConfigValue: NewAuthenticationObject(tt.accountName, tt.sasToken),
			}
			resp := &validator.ObjectResponse{}
			authenticationRequired{}.ValidateObject(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestGzipLevelCodecConflict(t *testing.T) {
	tests := []struct {
		name      string
		gzipLevel types.Int64
		codec     types.String
		wantError bool
	}{
		{
			name:      "gzip_level unset never conflicts",
			gzipLevel: types.Int64Null(),
			codec:     types.StringValue("gzip"),
			wantError: false,
		},
		{
			name:      "gzip_level set with empty codec is fine",
			gzipLevel: types.Int64Value(5),
			codec:     types.StringValue(""),
			wantError: false,
		},
		{
			name:      "gzip_level set with null codec is fine",
			gzipLevel: types.Int64Value(5),
			codec:     types.StringNull(),
			wantError: false,
		},
		{
			name:      "gzip_level set with non-empty codec conflicts",
			gzipLevel: types.Int64Value(5),
			codec:     types.StringValue("gzip"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := buildTestConfig(t, map[string]schema.Attribute{
				"gzip_level":        schema.Int64Attribute{Optional: true},
				"compression_codec": schema.StringAttribute{Optional: true},
			}, map[string]attr.Value{
				"gzip_level":        tt.gzipLevel,
				"compression_codec": tt.codec,
			})

			req := validator.Int64Request{
				Path:        path.Root("gzip_level"),
				ConfigValue: tt.gzipLevel,
				Config:      config,
			}
			resp := &validator.Int64Response{}

			gzipLevelCodecConflict{}.ValidateInt64(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestFileMaxBytesRange(t *testing.T) {
	tests := []struct {
		name      string
		value     types.Int64
		wantError bool
	}{
		{name: "unset", value: types.Int64Null(), wantError: false},
		{name: "zero means no limit", value: types.Int64Value(0), wantError: false},
		{name: "exactly the minimum", value: types.Int64Value(1048576), wantError: false},
		{name: "above the minimum", value: types.Int64Value(2097152), wantError: false},
		{name: "below the minimum and non-zero", value: types.Int64Value(1024), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.Int64Request{
				Path:        path.Root("file_max_bytes"),
				ConfigValue: tt.value,
			}
			resp := &validator.Int64Response{}

			fileMaxBytesRange{}.ValidateInt64(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestNotTrimmed(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{name: "unset", value: types.StringNull(), wantError: false},
		{name: "trimmed", value: types.StringValue("clean"), wantError: false},
		{name: "leading whitespace", value: types.StringValue(" clean"), wantError: true},
		{name: "trailing whitespace", value: types.StringValue("clean\n"), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("public_key"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			notTrimmed{}.ValidateString(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestValidateNoVCLOnlyAttributesForCompute(t *testing.T) {
	tests := []struct {
		name      string
		format    types.String
		version   types.Int64
		placement types.String
		respCond  types.String
		wantError bool
	}{
		{
			name:      "no VCL-only attributes configured",
			format:    types.StringNull(),
			version:   types.Int64Null(),
			placement: types.StringNull(),
			respCond:  types.StringNull(),
			wantError: false,
		},
		{
			name:      "format configured",
			format:    types.StringValue("custom-format"),
			version:   types.Int64Null(),
			placement: types.StringNull(),
			respCond:  types.StringNull(),
			wantError: true,
		},
		{
			name:      "all VCL-only attributes configured",
			format:    types.StringValue("custom-format"),
			version:   types.Int64Value(2),
			placement: types.StringValue("none"),
			respCond:  types.StringValue("cond"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := buildTestConfig(t, map[string]schema.Attribute{
				"format":             schema.StringAttribute{Optional: true},
				"format_version":     schema.Int64Attribute{Optional: true},
				"placement":          schema.StringAttribute{Optional: true},
				"response_condition": schema.StringAttribute{Optional: true},
			}, map[string]attr.Value{
				"format":             tt.format,
				"format_version":     tt.version,
				"placement":          tt.placement,
				"response_condition": tt.respCond,
			})

			result := ValidateNoVCLOnlyAttributesForCompute(context.Background(), cfg)

			assert.Equal(t, tt.wantError, result.HasError())
		})
	}
}

// buildTestConfig builds a minimal tfsdk.Config for validators that read
// sibling attributes via Config.GetAttribute. Only the attributes needed by
// the validator under test are included in the schema.
func buildTestConfig(t *testing.T, attrs map[string]schema.Attribute, values map[string]attr.Value) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	s := schema.Schema{Attributes: attrs}
	objType := s.Type().TerraformType(ctx)

	tfValues := make(map[string]tftypes.Value, len(values))
	for name, v := range values {
		tv, err := v.ToTerraformValue(ctx)
		if err != nil {
			t.Fatalf("building terraform value for %q: %v", name, err)
		}
		tfValues[name] = tv
	}

	raw := tftypes.NewValue(objType, tfValues)
	return tfsdk.Config{Raw: raw, Schema: s}
}
