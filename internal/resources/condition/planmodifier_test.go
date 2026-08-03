package condition

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeStatementModifier_PlanModifyString(t *testing.T) {
	tests := []struct {
		name     string
		state    types.String
		plan     types.String
		expected types.String
	}{
		{
			name:     "heredoc trailing newline against trimmed state keeps state value",
			state:    types.StringValue(`req.url ~ "^/admin"`),
			plan:     types.StringValue("req.url ~ \"^/admin\"\n"),
			expected: types.StringValue(`req.url ~ "^/admin"`),
		},
		{
			name:     "genuinely different statement keeps plan value",
			state:    types.StringValue(`req.url ~ "^/admin"`),
			plan:     types.StringValue(`req.url ~ "^/private"`),
			expected: types.StringValue(`req.url ~ "^/private"`),
		},
		{
			name:     "no prior state keeps plan value",
			state:    types.StringNull(),
			plan:     types.StringValue("req.url ~ \"^/admin\"\n"),
			expected: types.StringValue("req.url ~ \"^/admin\"\n"),
		},
		{
			name:     "unknown plan value is left unknown",
			state:    types.StringValue(`req.url ~ "^/admin"`),
			plan:     types.StringUnknown(),
			expected: types.StringUnknown(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue: tt.state,
				PlanValue:  tt.plan,
			}
			resp := &planmodifier.StringResponse{PlanValue: tt.plan}

			normalizeStatementModifier{}.PlanModifyString(context.Background(), req, resp)

			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}
