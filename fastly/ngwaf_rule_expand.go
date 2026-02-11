package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
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
		Type:               gofastly.ToPointer(d.Get("type").(string)),
		Description:        gofastly.ToPointer(d.Get("description").(string)),
		Scope:              s,
		Enabled:            gofastly.ToPointer(d.Get("enabled").(bool)),
		GroupOperator:      gofastly.ToPointer(d.Get("group_operator").(string)),
		RequestLogging:     gofastly.ToPointer(d.Get("request_logging").(string)),
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
		RuleID:             gofastly.ToPointer(d.Id()),
		Scope:              s,
		Type:               gofastly.ToPointer(d.Get("type").(string)),
		Description:        gofastly.ToPointer(d.Get("description").(string)),
		Enabled:            gofastly.ToPointer(d.Get("enabled").(bool)),
		GroupOperator:      gofastly.ToPointer(d.Get("group_operator").(string)),
		RequestLogging:     gofastly.ToPointer(d.Get("request_logging").(string)),
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
			Type: gofastly.ToPointer(m["type"].(string)),
		}
		if v, ok := m["signal"]; ok {
			action.Signal = gofastly.ToPointer(v.(string))
		}
		if scopeType == "workspace" {
			if v, ok := m["allow_interactive"]; ok && v.(bool) {
				action.AllowInteractive = gofastly.ToPointer(v.(bool))
			}
			if v, ok := m["deception_type"]; ok && v != "" {
				action.DeceptionType = gofastly.ToPointer(v.(string))
			}
			if v, ok := m["redirect_url"]; ok && v != "" {
				action.RedirectURL = gofastly.ToPointer(v.(string))
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
			Type: gofastly.ToPointer(m["type"].(string)),
		}
		if v, ok := m["signal"]; ok {
			action.Signal = gofastly.ToPointer(v.(string))
		}
		if scopeType == "workspace" {
			if v, ok := m["redirect_url"]; ok {
				action.RedirectURL = gofastly.ToPointer(v.(string))
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
			Field:    gofastly.ToPointer(field),
			Operator: gofastly.ToPointer(operator),
			Value:    gofastly.ToPointer(value),
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
			Field:    gofastly.ToPointer(field),
			Operator: gofastly.ToPointer(operator),
			Value:    gofastly.ToPointer(value),
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

		groupOp := m["group_operator"].(string)

		// Extract single conditions
		var conditions []*rules.CreateCondition
		if rawConditions, ok := m["condition"].([]any); ok {
			for _, c := range rawConditions {
				cm, ok := c.(map[string]any)
				if !ok {
					continue
				}
				conditions = append(conditions, &rules.CreateCondition{
					Field:    gofastly.ToPointer(cm["field"].(string)),
					Operator: gofastly.ToPointer(cm["operator"].(string)),
					Value:    gofastly.ToPointer(cm["value"].(string)),
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

				field := mvm["field"].(string)
				operator := mvm["operator"].(string)
				mvGroupOp := mvm["group_operator"].(string)

				// Extract nested conditions within the multival
				var mvConditions []*rules.CreateConditionMult
				if rawMVConds, ok := mvm["condition"].([]any); ok {
					for _, c := range rawMVConds {
						cm, ok := c.(map[string]any)
						if !ok {
							continue
						}
						mvConditions = append(mvConditions, &rules.CreateConditionMult{
							Field:    gofastly.ToPointer(cm["field"].(string)),
							Operator: gofastly.ToPointer(cm["operator"].(string)),
							Value:    gofastly.ToPointer(cm["value"].(string)),
						})
					}
				}

				multivalConditions = append(multivalConditions, &rules.CreateMultivalCondition{
					Field:         gofastly.ToPointer(field),
					Operator:      gofastly.ToPointer(operator),
					GroupOperator: gofastly.ToPointer(mvGroupOp),
					Conditions:    mvConditions,
				})
			}
		}

		groupConditions = append(groupConditions, &rules.CreateGroupCondition{
			GroupOperator:      gofastly.ToPointer(groupOp),
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

		groupOp := m["group_operator"].(string)

		// Extract single conditions
		var conditions []*rules.UpdateCondition
		if rawConditions, ok := m["condition"].([]any); ok {
			for _, c := range rawConditions {
				cm, ok := c.(map[string]any)
				if !ok {
					continue
				}
				conditions = append(conditions, &rules.UpdateCondition{
					Field:    gofastly.ToPointer(cm["field"].(string)),
					Operator: gofastly.ToPointer(cm["operator"].(string)),
					Value:    gofastly.ToPointer(cm["value"].(string)),
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

				field := mvm["field"].(string)
				operator := mvm["operator"].(string)
				mvGroupOp := mvm["group_operator"].(string)

				// Extract nested conditions within the multival
				var mvConditions []*rules.UpdateConditionMult
				if rawMVConds, ok := mvm["condition"].([]any); ok {
					for _, c := range rawMVConds {
						cm, ok := c.(map[string]any)
						if !ok {
							continue
						}
						mvConditions = append(mvConditions, &rules.UpdateConditionMult{
							Field:    gofastly.ToPointer(cm["field"].(string)),
							Operator: gofastly.ToPointer(cm["operator"].(string)),
							Value:    gofastly.ToPointer(cm["value"].(string)),
						})
					}
				}

				multivalConditions = append(multivalConditions, &rules.UpdateMultivalCondition{
					Field:         gofastly.ToPointer(field),
					Operator:      gofastly.ToPointer(operator),
					GroupOperator: gofastly.ToPointer(mvGroupOp),
					Conditions:    mvConditions,
				})
			}
		}

		groupConditions = append(groupConditions, &rules.UpdateGroupCondition{
			GroupOperator:      gofastly.ToPointer(groupOp),
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
				Field:    gofastly.ToPointer(cm["field"].(string)),
				Operator: gofastly.ToPointer(cm["operator"].(string)),
				Value:    gofastly.ToPointer(cm["value"].(string)),
			})
		}
		MultivalConditions = append(MultivalConditions, &rules.CreateMultivalCondition{
			Field:         gofastly.ToPointer(field),
			Operator:      gofastly.ToPointer(operator),
			GroupOperator: gofastly.ToPointer(groupOperator),
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
				Field:    gofastly.ToPointer(cm["field"].(string)),
				Operator: gofastly.ToPointer(cm["operator"].(string)),
				Value:    gofastly.ToPointer(cm["value"].(string)),
			})
		}
		MultivalConditions = append(MultivalConditions, &rules.UpdateMultivalCondition{
			Field:         gofastly.ToPointer(field),
			Operator:      gofastly.ToPointer(operator),
			GroupOperator: gofastly.ToPointer(groupOperator),
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
			Duration:          gofastly.ToPointer(m["duration"].(int)),
			Interval:          gofastly.ToPointer(m["interval"].(int)),
			Signal:            gofastly.ToPointer(m["signal"].(string)),
			Threshold:         gofastly.ToPointer(m["threshold"].(int)),
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
			Duration:          gofastly.ToPointer(m["duration"].(int)),
			Interval:          gofastly.ToPointer(m["interval"].(int)),
			Signal:            gofastly.ToPointer(m["signal"].(string)),
			Threshold:         gofastly.ToPointer(m["threshold"].(int)),
		}
	}

	return updateRateLimit
}
