package fastly

import (
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

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

func flattenNGWAFRuleRateLimitGeneric(rateLimit *rules.RateLimit) []map[string]any {
	clientIdentifiers := []map[string]any{}

	if rateLimit == nil {
		return clientIdentifiers
	}

	for _, ci := range rateLimit.ClientIdentifiers {
		m := map[string]any{}
		if ci.Key != "" {
			m["key"] = ci.Key
		}
		if ci.Name != "" {
			m["name"] = ci.Name
		}
		m["type"] = ci.Type

		clientIdentifiers = append(clientIdentifiers, m)
	}

	return []map[string]any{
		{
			"client_identifiers": clientIdentifiers,
			"duration":           rateLimit.Duration,
			"interval":           rateLimit.Interval,
			"signal":             rateLimit.Signal,
			"threshold":          rateLimit.Threshold,
		},
	}
}
