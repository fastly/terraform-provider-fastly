package loggingsplunk

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationRequired(t *testing.T) {
	tests := []struct {
		name      string
		token     types.String
		envToken  string
		wantError bool
	}{
		{
			name:      "token configured",
			token:     types.StringValue("splunk-token"),
			wantError: false,
		},
		{
			name:      "token unset, no env var",
			token:     types.StringNull(),
			wantError: true,
		},
		{
			name:      "token unset, env var set",
			token:     types.StringNull(),
			envToken:  "splunk-token",
			wantError: false,
		},
		{
			name:      "token explicitly blank is not rescued by env var",
			token:     types.StringValue(""),
			envToken:  "splunk-token",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(splunkTokenEnvVar, tt.envToken)

			req := validator.ObjectRequest{
				Path:        path.Root("authentication"),
				ConfigValue: NewAuthenticationObject(tt.token),
			}
			resp := &validator.ObjectResponse{}
			authenticationRequired{}.ValidateObject(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}

func TestAuthenticationRequired_BlockOmitted(t *testing.T) {
	tests := []struct {
		name      string
		envToken  string
		wantError bool
	}{
		{
			name:      "authentication block entirely omitted, no env var",
			wantError: true,
		},
		{
			name:      "authentication block entirely omitted, env var set",
			envToken:  "splunk-token",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(splunkTokenEnvVar, tt.envToken)

			req := validator.ObjectRequest{
				Path:        path.Root("authentication"),
				ConfigValue: types.ObjectNull(authenticationAttributeTypes),
			}
			resp := &validator.ObjectResponse{}
			authenticationRequired{}.ValidateObject(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}
