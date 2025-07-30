package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/common"
	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

func expandNGWAFRuleCreateInput(d *schema.ResourceData, scope *common.Scope) (*rules.CreateInput, error) {
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

	return &rules.CreateInput{
		Type:            gofastly.ToPointer(d.Get("type").(string)),
		Description:     gofastly.ToPointer(d.Get("description").(string)),
		Scope:           scope,
		Enabled:         gofastly.ToPointer(d.Get("enabled").(bool)),
		GroupOperator:   gofastly.ToPointer(d.Get("group_operator").(string)),
		RequestLogging:  gofastly.ToPointer(d.Get("request_logging").(string)),
		Actions:         expandNGWAFRuleCreateActions(actionRaw, string(scope.Type)),
		Conditions:      expandNGWAFRuleCreateConditions(conditionRaw),
		GroupConditions: expandNGWAFRuleGroupCreateConditions(groupRaw),
	}, nil
}

func expandNGWAFRuleUpdateInput(d *schema.ResourceData, scope *common.Scope) (*rules.UpdateInput, error) {
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

	return &rules.UpdateInput{
		RuleID:          gofastly.ToPointer(d.Id()),
		Scope:           scope,
		Type:            gofastly.ToPointer(d.Get("type").(string)),
		Description:     gofastly.ToPointer(d.Get("description").(string)),
		Enabled:         gofastly.ToPointer(d.Get("enabled").(bool)),
		GroupOperator:   gofastly.ToPointer(d.Get("group_operator").(string)),
		RequestLogging:  gofastly.ToPointer(d.Get("request_logging").(string)),
		Actions:         expandNGWAFRuleUpdateActions(actionRaw, string(scope.Type)),
		Conditions:      expandNGWAFRuleUpdateConditions(conditionRaw),
		GroupConditions: expandNGWAFRuleGroupUpdateConditions(groupRaw),
	}, nil
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
			Type:     gofastly.ToPointer("single"),
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
			Type:     gofastly.ToPointer("single"),
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

	groups := expandNGWAFRuleGroupConditionsGeneric(
		raw,
		func(field, operator, value string) any {
			return &rules.CreateCondition{
				Type:     gofastly.ToPointer("single"),
				Field:    gofastly.ToPointer(field),
				Operator: gofastly.ToPointer(operator),
				Value:    gofastly.ToPointer(value),
			}
		},
		func(groupOp string, conds []any) any {
			c := make([]*rules.CreateCondition, len(conds))
			for i, v := range conds {
				c[i] = v.(*rules.CreateCondition)
			}
			return &rules.CreateGroupCondition{
				Type:          gofastly.ToPointer("group"),
				GroupOperator: gofastly.ToPointer(groupOp),
				Conditions:    c,
			}
		},
	)
	result := make([]*rules.CreateGroupCondition, len(groups))
	for i, v := range groups {
		result[i] = v.(*rules.CreateGroupCondition)
	}

	return result
}

func expandNGWAFRuleGroupUpdateConditions(raw []any) []*rules.UpdateGroupCondition {
	if raw == nil {
		return nil
	}

	groups := expandNGWAFRuleGroupConditionsGeneric(
		raw,
		func(field, operator, value string) any {
			return &rules.UpdateCondition{
				Type:     gofastly.ToPointer("single"),
				Field:    gofastly.ToPointer(field),
				Operator: gofastly.ToPointer(operator),
				Value:    gofastly.ToPointer(value),
			}
		},
		func(groupOp string, conds []any) any {
			c := make([]*rules.UpdateCondition, len(conds))
			for i, v := range conds {
				c[i] = v.(*rules.UpdateCondition)
			}
			return &rules.UpdateGroupCondition{
				Type:          gofastly.ToPointer("group"),
				GroupOperator: gofastly.ToPointer(groupOp),
				Conditions:    c,
			}
		},
	)
	result := make([]*rules.UpdateGroupCondition, len(groups))
	for i, v := range groups {
		result[i] = v.(*rules.UpdateGroupCondition)
	}

	return result
}
