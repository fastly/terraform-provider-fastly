package fastly

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

func buildNGWAFRuleScope(d *schema.ResourceData) *common.Scope {
	if v, ok := d.GetOk("workspace_id"); ok {
		wsID := v.(string)
		if wsID != "" {
			return &common.Scope{
				Type:      common.ScopeTypeWorkspace,
				AppliesTo: []string{wsID},
			}
		}
	}

	if v, ok := d.GetOk("applies_to"); ok {
		rawList, ok := v.([]any)
		if !ok || len(rawList) == 0 {
			return nil
		}
		ids := make([]string, len(rawList))
		for i, id := range rawList {
			ids[i] = id.(string)
		}
		return &common.Scope{
			Type:      common.ScopeTypeAccount,
			AppliesTo: ids,
		}
	}

	return nil
}

func expandNGWAFRuleConditionsGeneric(
	raw []any,
	newFn func(field, operator, value string) any,
) []any {
	if raw == nil {
		return nil
	}

	var conditions []any
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		conditions = append(conditions, newFn(
			m["field"].(string),
			m["operator"].(string),
			m["value"].(string),
		))
	}

	return conditions
}

func expandNGWAFRuleGroupConditionsGeneric(
	raw []any,
	newFn func(field, operator, value string) any,
	groupFn func(operator string, conditions []any) any,
) []any {
	if raw == nil {
		return nil
	}

	var groupConditions []any
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		rawConditions, ok := m["condition"].([]any)
		if !ok || len(rawConditions) == 0 {
			continue
		}

		var conditions []any
		for _, c := range rawConditions {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			conditions = append(conditions, newFn(
				cm["field"].(string),
				cm["operator"].(string),
				cm["value"].(string),
			))
		}

		groupConditions = append(groupConditions, groupFn(m["group_operator"].(string), conditions))
	}

	return groupConditions
}

func flattenNGWAFRuleConditionsGeneric(items []rules.ConditionItem) ([]map[string]any, []map[string]any) {
	var singles []map[string]any
	var groups []map[string]any

	for _, item := range items {
		switch item.Type {
		case "single":
			if sc, ok := item.Fields.(rules.SingleCondition); ok {
				singles = append(singles, map[string]any{
					"field":    sc.Field,
					"operator": sc.Operator,
					"value":    sc.Value,
				})
			}

		case "group":
			if gc, ok := item.Fields.(rules.GroupCondition); ok {
				conds := make([]any, len(gc.Conditions))
				for i, c := range gc.Conditions {
					conds[i] = map[string]any{
						"field":    c.Field,
						"operator": c.Operator,
						"value":    c.Value,
					}
				}

				group := map[string]any{
					"group_operator": gc.GroupOperator,
					"condition":      conds,
				}

				groups = append(groups, group)
			}
		}
	}

	return singles, groups
}

func flattenNGWAFRuleActionsGeneric(actions []rules.Action, isWorkspace bool) []map[string]any {
	var result []map[string]any

	for _, a := range actions {
		action := map[string]any{
			"type": a.Type,
		}
		if a.Signal != "" {
			action["signal"] = a.Signal
		}
		if isWorkspace {
			if a.RedirectURL != "" {
				action["redirect_url"] = a.RedirectURL
			}
			if a.ResponseCode != 0 {
				action["response_code"] = a.ResponseCode
			}
		}
		result = append(result, action)
	}

	return result
}

func customNGWAFRuleImporter(scope common.ScopeType) *schema.ResourceImporter {
	return &schema.ResourceImporter{
		StateContext: func(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
			switch scope {
			case common.ScopeTypeWorkspace:
				// Expected format: "workspace_id/rule_id"
				parts := strings.SplitN(d.Id(), "/", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid ID format %q. Expected workspace_id/rule_id", d.Id())
				}
				workspaceID := parts[0]
				ruleID := parts[1]

				if err := d.Set("workspace_id", workspaceID); err != nil {
					return nil, fmt.Errorf("failed to set workspace_id: %w", err)
				}
				d.SetId(ruleID)

			case common.ScopeTypeAccount:
				// Only rule ID is needed for account-scoped rules
				if err := d.Set("applies_to", []string{"*"}); err != nil {
					return nil, fmt.Errorf("failed to set applies_to for account rule: %w", err)
				}
				d.SetId(d.Id())

			default:
				return nil, fmt.Errorf("unsupported scope type %q", scope)
			}

			return []*schema.ResourceData{d}, nil
		},
	}
}
