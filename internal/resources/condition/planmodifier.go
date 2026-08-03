package condition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type normalizeStatementModifier struct{}

func (m normalizeStatementModifier) Description(_ context.Context) string {
	return "Suppresses diffs against the prior state when the only difference in `statement` is leading/trailing whitespace, e.g. a trailing newline from a HEREDOC."
}

func (m normalizeStatementModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeStatementModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if trimStatement(req.StateValue.ValueString()) == trimStatement(req.PlanValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}
