package fastly

import (
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
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

func flattenNGWAFRuleConditionsGeneric(items []rules.ConditionItem) ([]map[string]any, []map[string]any, []map[string]any) {
	var singles []map[string]any
	var groups []map[string]any
	var multivals []map[string]any

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
				var conds []any
				var multivalConds []any

				// Iterate through all conditions in the group
				for _, gci := range gc.Conditions {
					switch gci.Type {
					case "single":
						if sc, ok := gci.Fields.(rules.Condition); ok {
							conds = append(conds, map[string]any{
								"field":    sc.Field,
								"operator": sc.Operator,
								"value":    sc.Value,
							})
						}

					case "multival":
						if mc, ok := gci.Fields.(rules.MultivalCondition); ok {
							mvConds := make([]any, len(mc.Conditions))
							for i, c := range mc.Conditions {
								mvConds[i] = map[string]any{
									"field":    c.Field,
									"operator": c.Operator,
									"value":    c.Value,
								}
							}

							multivalConds = append(multivalConds, map[string]any{
								"field":          mc.Field,
								"operator":       mc.Operator,
								"group_operator": mc.GroupOperator,
								"condition":      mvConds,
							})
						}
					}
				}

				group := map[string]any{
					"group_operator": gc.GroupOperator,
				}
				if len(conds) > 0 {
					group["condition"] = conds
				}
				if len(multivalConds) > 0 {
					group["multival_condition"] = multivalConds
				}

				groups = append(groups, group)
			}

		case "multival":
			if mc, ok := item.Fields.(rules.MultivalCondition); ok {
				conds := make([]any, len(mc.Conditions))
				for i, c := range mc.Conditions {
					conds[i] = map[string]any{
						"field":    c.Field,
						"operator": c.Operator,
						"value":    c.Value,
					}
				}

				multival := map[string]any{
					"field":          mc.Field,
					"operator":       mc.Operator,
					"group_operator": mc.GroupOperator,
					"condition":      conds,
				}

				multivals = append(multivals, multival)
			}
		}
	}

	return singles, groups, multivals
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
			if a.AllowInteractive != nil {
				action["allow_interactive"] = a.AllowInteractive
			}
			if a.DeceptionType != "" {
				action["deception_type"] = a.DeceptionType
			}
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
