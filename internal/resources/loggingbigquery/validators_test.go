package loggingbigquery

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

func TestAuthenticationEitherOr(t *testing.T) {
	tests := []struct {
		name                 string
		accountName          types.String
		email                types.String
		secretKey            types.String
		envAccount           string
		envDeprecatedAccount string
		envEmail             string
		envSecret            string
		wantError            bool
	}{
		{
			name:        "account_name only",
			accountName: types.StringValue("svc-account"),
			email:       types.StringNull(),
			secretKey:   types.StringNull(),
			wantError:   false,
		},
		{
			name:        "email and secret_key",
			accountName: types.StringNull(),
			email:       types.StringValue("a@b.com"),
			secretKey:   types.StringValue("secret"),
			wantError:   false,
		},
		{
			name:        "all three set",
			accountName: types.StringValue("svc-account"),
			email:       types.StringValue("a@b.com"),
			secretKey:   types.StringValue("secret"),
			wantError:   false,
		},
		{
			name:        "email only",
			accountName: types.StringNull(),
			email:       types.StringValue("a@b.com"),
			secretKey:   types.StringNull(),
			wantError:   true,
		},
		{
			name:        "secret_key only",
			accountName: types.StringNull(),
			email:       types.StringNull(),
			secretKey:   types.StringValue("secret"),
			wantError:   true,
		},
		{
			name:        "nothing configured, no env vars",
			accountName: types.StringNull(),
			email:       types.StringNull(),
			secretKey:   types.StringNull(),
			wantError:   true,
		},
		{
			name:        "nothing configured, account_name env var set",
			accountName: types.StringNull(),
			email:       types.StringNull(),
			secretKey:   types.StringNull(),
			envAccount:  "svc-account",
			wantError:   false,
		},
		{
			name:                 "nothing configured, deprecated account_name env var set",
			accountName:          types.StringNull(),
			email:                types.StringNull(),
			secretKey:            types.StringNull(),
			envDeprecatedAccount: "legacy-svc-account",
			wantError:            false,
		},
		{
			name:        "email configured, secret_key falls back to env var",
			accountName: types.StringNull(),
			email:       types.StringValue("a@b.com"),
			secretKey:   types.StringNull(),
			envSecret:   "secret",
			wantError:   false,
		},
		{
			name:        "email configured, secret_key unset and no env var",
			accountName: types.StringNull(),
			email:       types.StringValue("a@b.com"),
			secretKey:   types.StringNull(),
			wantError:   true,
		},
		{
			// account_name explicitly "" is configured, not omitted — Terraform's
			// schema Default never overrides it with the env var, so the validator
			// must not treat it as satisfying auth just because the env var is set.
			name:        "account_name explicitly blank does not fall back to env var",
			accountName: types.StringValue(""),
			email:       types.StringNull(),
			secretKey:   types.StringNull(),
			envAccount:  "svc-account",
			wantError:   true,
		},
		{
			// Same as above for email/secret_key: an explicit "" is configured, so
			// it must not fall back to the env var either.
			name:        "email explicitly blank does not fall back to env var",
			accountName: types.StringNull(),
			email:       types.StringValue(""),
			secretKey:   types.StringValue("secret"),
			envEmail:    "a@b.com",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME", tt.envAccount)
			t.Setenv("FASTLY_GCS_ACCOUNT_NAME", tt.envDeprecatedAccount)
			t.Setenv("FASTLY_BQ_EMAIL", tt.envEmail)
			t.Setenv("FASTLY_BQ_SECRET_KEY", tt.envSecret)

			req := validator.ObjectRequest{
				Path:        path.Root("authentication"),
				ConfigValue: NewAuthenticationObject(tt.accountName, tt.email, tt.secretKey),
			}
			resp := &validator.ObjectResponse{}
			authenticationEitherOr{}.ValidateObject(context.Background(), req, resp)

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
			name:      "format_version configured",
			format:    types.StringNull(),
			version:   types.Int64Value(2),
			placement: types.StringNull(),
			respCond:  types.StringNull(),
			wantError: true,
		},
		{
			name:      "placement configured",
			format:    types.StringNull(),
			version:   types.Int64Null(),
			placement: types.StringValue("none"),
			respCond:  types.StringNull(),
			wantError: true,
		},
		{
			name:      "response_condition configured",
			format:    types.StringNull(),
			version:   types.Int64Null(),
			placement: types.StringNull(),
			respCond:  types.StringValue("cond"),
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
