package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v17/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v17/fastly/ngwaf/v1/scope"
)

func expandNGWAFRuleCreateInput(d *schema.ResourceData, s *scope.Scope) *rules.CreateInput {
	var actionRaw []any
	if v, ok := d.GetOk("action"); ok {
		actionRaw = v.([]any)
	}

	var conditionRaw []any
	if v, ok := d.GetOk("condition"); ok {
		conditionRaw = v.([]any)
	}

	var groupRaw []any
	if v, ok := d.GetOk("group_condition"); ok {
		groupRaw = v.([]any)
	}

	var multivalRaw []any
	if v, ok := d.GetOk("multival_condition"); ok {
		multivalRaw = v.([]any)
	}

	var rateLimitRaw []any
	if v, ok := d.GetOk("rate_limit"); ok {
		rateLimitRaw = v.([]any)
	}

	return &rules.CreateInput{
		Type:               new(d.Get("type").(string)),
		Description:        new(d.Get("description").(string)),
		Scope:              s,
		Enabled:            new(d.Get("enabled").(bool)),
		GroupOperator:      new(d.Get("group_operator").(string)),
		RequestLogging:     new(d.Get("request_logging").(string)),
		Actions:            expandNGWAFRuleCreateActions(actionRaw, string(s.Type)),
		Conditions:         expandNGWAFRuleCreateConditions(conditionRaw),
		GroupConditions:    expandNGWAFRuleGroupCreateConditions(groupRaw),
		MultivalConditions: expandNGWAFRuleMultiValCreateConditions(multivalRaw),
		RateLimit:          expandNGWAFRuleCreateRateLimit(rateLimitRaw),
	}
}

func expandNGWAFRuleUpdateInput(d *schema.ResourceData, s *scope.Scope) *rules.UpdateInput {
	var actionRaw []any
	if v, ok := d.GetOk("action"); ok {
		actionRaw = v.([]any)
	}

	var conditionRaw []any
	if v, ok := d.GetOk("condition"); ok {
		conditionRaw = v.([]any)
	}

	var groupRaw []any
	if v, ok := d.GetOk("group_condition"); ok {
		groupRaw = v.([]any)
	}

	var multivalRaw []any
	if v, ok := d.GetOk("multival_condition"); ok {
		multivalRaw = v.([]any)
	}

	var rateLimitRaw []any
	if v, ok := d.GetOk("rate_limit"); ok {
		rateLimitRaw = v.([]any)
	}

	updateInput := &rules.UpdateInput{
		RuleID:             new(d.Id()),
		Scope:              s,
		Type:               new(d.Get("type").(string)),
		Description:        new(d.Get("description").(string)),
		Enabled:            new(d.Get("enabled").(bool)),
		GroupOperator:      new(d.Get("group_operator").(string)),
		RequestLogging:     new(d.Get("request_logging").(string)),
		Conditions:         expandNGWAFRuleUpdateConditions(conditionRaw),
		GroupConditions:    expandNGWAFRuleGroupUpdateConditions(groupRaw),
		MultivalConditions: expandNGWAFRuleMultiValUpdateConditions(multivalRaw),
		RateLimit:          expandNGWAFRuleUpdateRateLimit(rateLimitRaw),
	}

	// templated_signal rules don't allow actions in update requests
	if d.Get("type").(string) != "templated_signal" {
		updateInput.Actions = expandNGWAFRuleUpdateActions(actionRaw, string(s.Type))
	}

	return updateInput
}

func expandNGWAFRuleCreateActions(raw []any, scopeType string) []*rules.CreateAction {
	if raw == nil {
		return nil
	}

	var actions []*rules.CreateAction
	for _, item := range raw {
		m := item.(map[string]any)
		action := &rules.CreateAction{
			Type: new(m["type"].(string)),
		}
		if v, ok := m["signal"]; ok {
			action.Signal = new(v.(string))
		}
		if scopeType == "workspace" {
			if v, ok := m["allow_interactive"]; ok && v.(bool) {
				action.AllowInteractive = new(v.(bool))
			}
			if v, ok := m["deception_type"]; ok && v != "" {
				action.DeceptionType = new(v.(string))
			}
			if v, ok := m["redirect_url"]; ok && v != "" {
				action.RedirectURL = new(v.(string))
			}
			if v, ok := m["response_code"]; ok && v != 0 {
				val := v.(int)
				action.ResponseCode = &val
			}
		}
		actions = append(actions, action)
	}

	return actions
}

func expandNGWAFRuleUpdateActions(raw []any, scopeType string) []*rules.UpdateAction {
	if raw == nil {
		return nil
	}

	var actions []*rules.UpdateAction
	for _, item := range raw {
		m := item.(map[string]any)
		action := &rules.UpdateAction{
			Type: new(m["type"].(string)),
		}
		if v, ok := m["signal"]; ok {
			action.Signal = new(v.(string))
		}
		if scopeType == "workspace" {
			if v, ok := m["allow_interactive"]; ok && v.(bool) {
				action.AllowInteractive = new(v.(bool))
			}
			if v, ok := m["deception_type"]; ok && v != "" {
				action.DeceptionType = new(v.(string))
			}
			if v, ok := m["redirect_url"]; ok {
				action.RedirectURL = new(v.(string))
			}
			if v, ok := m["response_code"]; ok {
				val := v.(int)
				action.ResponseCode = &val
			}
		}
		actions = append(actions, action)
	}

	return actions
}

func expandNGWAFRuleCreateConditions(raw []any) []*rules.CreateCondition {
	if raw == nil {
		return nil
	}

	conds := expandNGWAFRuleConditionsGeneric(raw, func(field, operator, value string) any {
		return &rules.CreateCondition{
			Field:    new(field),
			Operator: new(operator),
			Value:    new(value),
		}
	})
	result := make([]*rules.CreateCondition, len(conds))
	for i, v := range conds {
		result[i] = v.(*rules.CreateCondition)
	}

	return result
}

func expandNGWAFRuleUpdateConditions(raw []any) []*rules.UpdateCondition {
	if raw == nil {
		return nil
	}

	conds := expandNGWAFRuleConditionsGeneric(raw, func(field, operator, value string) any {
		return &rules.UpdateCondition{
			Field:    new(field),
			Operator: new(operator),
			Value:    new(value),
		}
	})
	result := make([]*rules.UpdateCondition, len(conds))
	for i, v := range conds {
		result[i] = v.(*rules.UpdateCondition)
	}

	return result
}

func expandNGWAFRuleGroupCreateConditions(raw []any) []*rules.CreateGroupCondition {
	if raw == nil {
		return nil
	}

	var groupConditions []*rules.CreateGroupCondition
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		groupOp, ok := m["group_operator"].(string)
		if !ok {
			continue
		}

		// Extract single conditions
		var conditions []*rules.CreateCondition
		if rawConditions, ok := m["condition"].([]any); ok {
			for _, c := range rawConditions {
				cm, ok := c.(map[string]any)
				if !ok {
					continue
				}
				field, ok := cm["field"].(string)
				if !ok {
					continue
				}
				operator, ok := cm["operator"].(string)
				if !ok {
					continue
				}
				value, ok := cm["value"].(string)
				if !ok {
					continue
				}
				conditions = append(conditions, &rules.CreateCondition{
					Field:    new(field),
					Operator: new(operator),
					Value:    new(value),
				})
			}
		}

		// Extract multival conditions
		var multivalConditions []*rules.CreateMultivalCondition
		if rawMultivals, ok := m["multival_condition"].([]any); ok {
			for _, mv := range rawMultivals {
				mvm, ok := mv.(map[string]any)
				if !ok {
					continue
				}

				field, ok := mvm["field"].(string)
				if !ok {
					continue
				}
				operator, ok := mvm["operator"].(string)
				if !ok {
					continue
				}
				mvGroupOp, ok := mvm["group_operator"].(string)
				if !ok {
					continue
				}

				// Extract nested conditions within the multival
				var mvConditions []*rules.CreateConditionMult
				if rawMVConds, ok := mvm["condition"].([]any); ok {
					for _, c := range rawMVConds {
						cm, ok := c.(map[string]any)
						if !ok {
							continue
						}
						mvField, ok := cm["field"].(string)
						if !ok {
							continue
						}
						mvOperator, ok := cm["operator"].(string)
						if !ok {
							continue
						}
						mvValue, ok := cm["value"].(string)
						if !ok {
							continue
						}
						mvConditions = append(mvConditions, &rules.CreateConditionMult{
							Field:    new(mvField),
							Operator: new(mvOperator),
							Value:    new(mvValue),
						})
					}
				}

				multivalConditions = append(multivalConditions, &rules.CreateMultivalCondition{
					Field:         new(field),
					Operator:      new(operator),
					GroupOperator: new(mvGroupOp),
					Conditions:    mvConditions,
				})
			}
		}

		groupConditions = append(groupConditions, &rules.CreateGroupCondition{
			GroupOperator:      new(groupOp),
			Conditions:         conditions,
			MultivalConditions: multivalConditions,
		})
	}

	return groupConditions
}

func expandNGWAFRuleGroupUpdateConditions(raw []any) []*rules.UpdateGroupCondition {
	if raw == nil {
		return nil
	}

	var groupConditions []*rules.UpdateGroupCondition
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		groupOp, ok := m["group_operator"].(string)
		if !ok {
			continue
		}

		// Extract single conditions
		var conditions []*rules.UpdateCondition
		if rawConditions, ok := m["condition"].([]any); ok {
			for _, c := range rawConditions {
				cm, ok := c.(map[string]any)
				if !ok {
					continue
				}
				field, ok := cm["field"].(string)
				if !ok {
					continue
				}
				operator, ok := cm["operator"].(string)
				if !ok {
					continue
				}
				value, ok := cm["value"].(string)
				if !ok {
					continue
				}
				conditions = append(conditions, &rules.UpdateCondition{
					Field:    new(field),
					Operator: new(operator),
					Value:    new(value),
				})
			}
		}

		// Extract multival conditions
		var multivalConditions []*rules.UpdateMultivalCondition
		if rawMultivals, ok := m["multival_condition"].([]any); ok {
			for _, mv := range rawMultivals {
				mvm, ok := mv.(map[string]any)
				if !ok {
					continue
				}

				field, ok := mvm["field"].(string)
				if !ok {
					continue
				}
				operator, ok := mvm["operator"].(string)
				if !ok {
					continue
				}
				mvGroupOp, ok := mvm["group_operator"].(string)
				if !ok {
					continue
				}

				// Extract nested conditions within the multival
				var mvConditions []*rules.UpdateConditionMult
				if rawMVConds, ok := mvm["condition"].([]any); ok {
					for _, c := range rawMVConds {
						cm, ok := c.(map[string]any)
						if !ok {
							continue
						}
						mvField, ok := cm["field"].(string)
						if !ok {
							continue
						}
						mvOperator, ok := cm["operator"].(string)
						if !ok {
							continue
						}
						mvValue, ok := cm["value"].(string)
						if !ok {
							continue
						}
						mvConditions = append(mvConditions, &rules.UpdateConditionMult{
							Field:    new(mvField),
							Operator: new(mvOperator),
							Value:    new(mvValue),
						})
					}
				}

				multivalConditions = append(multivalConditions, &rules.UpdateMultivalCondition{
					Field:         new(field),
					Operator:      new(operator),
					GroupOperator: new(mvGroupOp),
					Conditions:    mvConditions,
				})
			}
		}

		groupConditions = append(groupConditions, &rules.UpdateGroupCondition{
			GroupOperator:      new(groupOp),
			Conditions:         conditions,
			MultivalConditions: multivalConditions,
		})
	}

	return groupConditions
}

func expandNGWAFRuleMultiValCreateConditions(raw []any) []*rules.CreateMultivalCondition {
	if raw == nil {
		return nil
	}

	var MultivalConditions []*rules.CreateMultivalCondition
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// Extract top level fields.
		field := m["field"].(string)
		operator := m["operator"].(string)
		groupOperator := m["group_operator"].(string)

		// Extract nested conditions.
		rawConditions, ok := m["condition"].([]any)
		if !ok || len(rawConditions) == 0 {
			continue
		}

		var conditions []*rules.CreateConditionMult
		for _, c := range rawConditions {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			conditions = append(conditions, &rules.CreateConditionMult{
				Field:    new(cm["field"].(string)),
				Operator: new(cm["operator"].(string)),
				Value:    new(cm["value"].(string)),
			})
		}
		MultivalConditions = append(MultivalConditions, &rules.CreateMultivalCondition{
			Field:         new(field),
			Operator:      new(operator),
			GroupOperator: new(groupOperator),
			Conditions:    conditions,
		})
	}

	return MultivalConditions
}

func expandNGWAFRuleMultiValUpdateConditions(raw []any) []*rules.UpdateMultivalCondition {
	if raw == nil {
		return nil
	}

	var MultivalConditions []*rules.UpdateMultivalCondition
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// Extract top level fields.
		field := m["field"].(string)
		operator := m["operator"].(string)
		groupOperator := m["group_operator"].(string)

		// Extract nested conditions.
		rawConditions, ok := m["condition"].([]any)
		if !ok || len(rawConditions) == 0 {
			continue
		}

		var conditions []*rules.UpdateConditionMult
		for _, c := range rawConditions {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			conditions = append(conditions, &rules.UpdateConditionMult{
				Field:    new(cm["field"].(string)),
				Operator: new(cm["operator"].(string)),
				Value:    new(cm["value"].(string)),
			})
		}
		MultivalConditions = append(MultivalConditions, &rules.UpdateMultivalCondition{
			Field:         new(field),
			Operator:      new(operator),
			GroupOperator: new(groupOperator),
			Conditions:    conditions,
		})
	}

	return MultivalConditions
}

func expandNGWAFRuleCreateRateLimit(raw []any) *rules.CreateRateLimit {
	if raw == nil {
		return nil
	}

	genericElement := raw[0]
	castElement := genericElement.(map[string]any)

	var createRateLimitClientIdentifiers []*rules.CreateClientIdentifier
	for _, m := range castElement["client_identifiers"].(*schema.Set).List() {
		key := m.(map[string]any)["key"].(string)
		name := m.(map[string]any)["name"].(string)
		t := m.(map[string]any)["type"].(string)

		ci := rules.CreateClientIdentifier{
			Key:  &key,
			Name: &name,
			Type: &t,
		}

		createRateLimitClientIdentifiers = append(createRateLimitClientIdentifiers, &ci)
	}

	var createRateLimit *rules.CreateRateLimit
	for _, item := range raw {
		m := item.(map[string]any)
		createRateLimit = &rules.CreateRateLimit{
			ClientIdentifiers: createRateLimitClientIdentifiers,
			Duration:          new(m["duration"].(int)),
			Interval:          new(m["interval"].(int)),
			Signal:            new(m["signal"].(string)),
			Threshold:         new(m["threshold"].(int)),
		}
	}

	return createRateLimit
}

func expandNGWAFRuleUpdateRateLimit(raw []any) *rules.UpdateRateLimit {
	if raw == nil {
		return nil
	}

	genericElement := raw[0]
	castElement := genericElement.(map[string]any)

	var updateRateLimitClientIdentifiers []*rules.UpdateClientIdentifier
	for _, m := range castElement["client_identifiers"].(*schema.Set).List() {
		key := m.(map[string]any)["key"].(string)
		name := m.(map[string]any)["name"].(string)
		t := m.(map[string]any)["type"].(string)

		ci := rules.UpdateClientIdentifier{
			Key:  &key,
			Name: &name,
			Type: &t,
		}

		updateRateLimitClientIdentifiers = append(updateRateLimitClientIdentifiers, &ci)
	}

	var updateRateLimit *rules.UpdateRateLimit
	for _, item := range raw {
		m := item.(map[string]any)
		updateRateLimit = &rules.UpdateRateLimit{
			ClientIdentifiers: updateRateLimitClientIdentifiers,
			Duration:          new(m["duration"].(int)),
			Interval:          new(m["interval"].(int)),
			Signal:            new(m["signal"].(string)),
			Threshold:         new(m["threshold"].(int)),
		}
	}

	return updateRateLimit
}
