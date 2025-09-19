package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

func flattenNGWAFRuleResponse(d *schema.ResourceData, rule *rules.Rule) error {
	if rule == nil {
		return fmt.Errorf("cannot flatten nil rule")
	}
	if rule.Scope.Type == "" || len(rule.Scope.AppliesTo) == 0 {
		return fmt.Errorf("invalid rule scope: type or applies_to is missing")
	}

	s := rule.Scope

	switch s.Type {
	case string(scope.ScopeTypeWorkspace):
		if len(s.AppliesTo) == 0 {
			return fmt.Errorf("workspace scope is missing applies_to ID")
		}
		if err := d.Set("workspace_id", s.AppliesTo[0]); err != nil {
			return fmt.Errorf("error setting workspace_id: %w", err)
		}

	case string(scope.ScopeTypeAccount):
		if err := d.Set("applies_to", s.AppliesTo); err != nil {
			return fmt.Errorf("error setting applies_to: %w", err)
		}

	default:
		return fmt.Errorf("unknown scope type: %q", s.Type)
	}

	if err := d.Set("description", rule.Description); err != nil {
		return fmt.Errorf("error setting description: %w", err)
	}

	if err := d.Set("type", rule.Type); err != nil {
		return fmt.Errorf("error setting type: %w", err)
	}

	if err := d.Set("enabled", rule.Enabled); err != nil {
		return fmt.Errorf("error setting enabled: %w", err)
	}

	if err := d.Set("group_operator", rule.GroupOperator); err != nil {
		return fmt.Errorf("error setting group_operator: %w", err)
	}

	if err := d.Set("request_logging", rule.RequestLogging); err != nil {
		return fmt.Errorf("error setting request_logging: %w", err)
	}

	isWorkspace := s.Type == string(scope.ScopeTypeWorkspace)

	// Flatten actions
	if err := d.Set("action", flattenNGWAFRuleActionsGeneric(rule.Actions, isWorkspace)); err != nil {
		return fmt.Errorf("error setting actions: %w", err)
	}

	// Flatten conditions
	singles, groups, multivals := flattenNGWAFRuleConditionsGeneric(rule.Conditions)

	if err := d.Set("condition", singles); err != nil {
		return fmt.Errorf("error setting condition: %w", err)
	}

	if err := d.Set("group_condition", groups); err != nil {
		return fmt.Errorf("error setting group_condition: %w", err)
	}

	if err := d.Set("multival_condition", multivals); err != nil {
		return fmt.Errorf("error setting multival_condition: %w", err)
	}

	// Flatten rate limit
	if isWorkspace {
		if err := d.Set("rate_limit", flattenNGWAFRuleRateLimitGeneric(rule.RateLimit)); err != nil {
			return fmt.Errorf("error setting rate_limit: %w", err)
		}
	}

	return nil
}
