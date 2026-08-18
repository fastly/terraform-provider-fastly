package logginghttps

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestGzipLevelCodecConflict(t *testing.T) {
	tests := []struct {
		name      string
		gzipLevel types.Int64
		codec     types.String
		wantError bool
	}{
		{"unset gzip_level is fine with a codec", types.Int64Null(), types.StringValue("gzip"), false},
		{"gzip_level with blank codec is fine", types.Int64Value(5), types.StringValue(""), false},
		{"gzip_level with null codec is fine", types.Int64Value(5), types.StringNull(), false},
		{"gzip_level with a codec conflicts", types.Int64Value(5), types.StringValue("gzip"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := resourceschema.Schema{Attributes: CommonAttributes()}

			resp := &validator.Int64Response{}
			gzipLevelCodecConflict{}.ValidateInt64(context.Background(), validator.Int64Request{
				Path:        path.Root("gzip_level"),
				ConfigValue: tt.gzipLevel,
				Config:      fakeConfigWithCodec(t, s, tt.codec),
			}, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

// fakeConfigWithCodec builds a tfsdk.Config whose compression_codec attribute
// resolves to codec, for exercising gzipLevelCodecConflict's cross-attribute
// lookup without a full plan/apply cycle.
func fakeConfigWithCodec(t *testing.T, s resourceschema.Schema, codec types.String) tfsdk.Config {
	t.Helper()

	vals := map[string]tftypes.Value{}
	tfType := s.Type().TerraformType(context.Background())
	for name, attrType := range tfType.(tftypes.Object).AttributeTypes {
		if name == "compression_codec" {
			if codec.IsNull() {
				vals[name] = tftypes.NewValue(attrType, nil)
			} else {
				vals[name] = tftypes.NewValue(attrType, codec.ValueString())
			}
			continue
		}
		vals[name] = tftypes.NewValue(attrType, nil)
	}

	return tfsdk.Config{
		Raw:    tftypes.NewValue(tfType, vals),
		Schema: s,
	}
}

func TestHTTPSURL(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid https URL", "https://example.com/logs", false},
		{"http scheme rejected", "http://example.com/logs", true},
		{"missing host", "https://", true},
		{"not a URL", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &validator.StringResponse{}
			httpsURL{}.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("url"),
				ConfigValue: types.StringValue(tt.value),
			}, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestValidateNoVCLOnlyAttributesForCompute(t *testing.T) {
	s := resourceschema.Schema{Attributes: CommonAttributes()}

	tests := []struct {
		name      string
		format    types.String
		placement types.String
		wantError bool
	}{
		{"no VCL-only attributes configured", types.StringNull(), types.StringNull(), false},
		{"format configured", types.StringValue("some-format"), types.StringNull(), true},
		{"placement configured", types.StringNull(), types.StringValue("none"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := map[string]tftypes.Value{}
			tfType := s.Type().TerraformType(context.Background())
			for name, attrType := range tfType.(tftypes.Object).AttributeTypes {
				switch name {
				case "format":
					if tt.format.IsNull() {
						vals[name] = tftypes.NewValue(attrType, nil)
					} else {
						vals[name] = tftypes.NewValue(attrType, tt.format.ValueString())
					}
				case "placement":
					if tt.placement.IsNull() {
						vals[name] = tftypes.NewValue(attrType, nil)
					} else {
						vals[name] = tftypes.NewValue(attrType, tt.placement.ValueString())
					}
				default:
					vals[name] = tftypes.NewValue(attrType, nil)
				}
			}

			cfg := tfsdk.Config{Raw: tftypes.NewValue(tfType, vals), Schema: s}
			diags := ValidateNoVCLOnlyAttributesForCompute(context.Background(), cfg)
			require.Equal(t, tt.wantError, diags.HasError())
		})
	}
}
