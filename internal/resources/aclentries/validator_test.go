package aclentries

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestValidEntriesValidator(t *testing.T) {
	tests := []struct {
		name      string
		entries   map[string]attr.Value
		wantError bool
	}{
		{
			name: "valid entries",
			entries: map[string]attr.Value{
				"192.0.2.0/24": types.StringValue("ALLOW"),
			},
		},
		{
			name: "invalid CIDR prefix",
			entries: map[string]attr.Value{
				"not_a_cidr": types.StringValue("ALLOW"),
			},
			wantError: true,
		},
		{
			name: "invalid action",
			entries: map[string]attr.Value{
				"192.0.2.0/24": types.StringValue("PERMIT"),
			},
			wantError: true,
		},
		{
			name: "unknown value is skipped for action check",
			entries: map[string]attr.Value{
				"192.0.2.0/24": types.StringUnknown(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, diags := types.MapValue(types.StringType, tt.entries)
			if diags.HasError() {
				t.Fatalf("unexpected error building map: %v", diags)
			}

			req := validator.MapRequest{
				Path:        path.Root("entries"),
				ConfigValue: m,
			}
			resp := &validator.MapResponse{}

			ValidEntries().ValidateMap(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}
