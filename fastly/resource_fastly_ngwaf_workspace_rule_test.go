package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v11/fastly/ngwaf/v1/rules"
)

func TestFlattenNGWAFRuleResponse(t *testing.T) {
	schemaMap := resourceFastlyNGWAFWorkspaceRule().Schema
	d := schema.TestResourceDataRaw(t, schemaMap, map[string]any{})

	rule := &rules.Rule{
		RuleID:         "example-rule-id",
		Type:           "request",
		Description:    "Terraform Rule Unit Test",
		Enabled:        true,
		GroupOperator:  "all",
		RequestLogging: "sampled",
		Scope: rules.Scope{
			Type:      "workspace",
			AppliesTo: []string{"workspace-123"},
		},
		Actions: []rules.Action{
			{Type: "block"},
		},
		Conditions: []rules.ConditionItem{
			{
				Type:   "single",
				Fields: rules.SingleCondition{Field: "ip", Operator: "equals", Value: "127.0.0.1"},
			},
			{
				Type:   "single",
				Fields: rules.SingleCondition{Field: "path", Operator: "equals", Value: "/login"},
			},
			{
				Type:   "single",
				Fields: rules.SingleCondition{Field: "agent_name", Operator: "equals", Value: "host-001"},
			},
			{
				Type: "group",
				Fields: rules.GroupCondition{
					GroupOperator: "all",
					Conditions: []rules.Condition{
						{Type: "single", Field: "country", Operator: "equals", Value: "AD"},
						{Type: "single", Field: "method", Operator: "equals", Value: "POST"},
					},
				},
			},
			{
				Type: "group",
				Fields: rules.GroupCondition{
					GroupOperator: "any",
					Conditions: []rules.Condition{
						{Type: "single", Field: "protocol_version", Operator: "equals", Value: "HTTP/1.0"},
						{Type: "single", Field: "method", Operator: "equals", Value: "HEAD"},
						{Type: "single", Field: "domain", Operator: "equals", Value: "example.com"},
					},
				},
			},
		},
	}

	require.NoError(t, flattenNGWAFRuleResponse(d, rule), "flattenNGWAFRuleResponse should not error")

	// Simple value checks
	require.Equal(t, "Terraform Rule Unit Test", d.Get("description"))
	require.Equal(t, "request", d.Get("type"))
	require.Equal(t, "all", d.Get("group_operator"))
	require.Equal(t, "sampled", d.Get("request_logging"))
	require.Equal(t, "workspace-123", d.Get("workspace_id"))

	// Action
	actions := d.Get("action").([]any)
	require.Len(t, actions, 1)
	require.Equal(t, "block", actions[0].(map[string]any)["type"])

	// Conditions
	conds := d.Get("condition").([]any)
	require.Len(t, conds, 3)

	// Group conditions
	groups := d.Get("group_condition").([]any)
	require.Len(t, groups, 2)
}

func TestAccFastlyNGWAFWorkspaceRule_basic(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	ruleDescription := fmt.Sprintf("Terraform Rule %s", acctest.RandString(5))
	updatedRuleDescription := ruleDescription + " updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil, // Rule is deleted implicitly when workspace is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceRuleConfig(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "type", "request"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "request_logging", "sampled"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_operator", "all"),

					// Action
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "action.0.type", "block"),

					// Conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.value", "127.0.0.1"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.value", "/login"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.field", "agent_name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.value", "host-001"),

					// Group conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.field", "country"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.value", "AD"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.field", "method"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.value", "POST"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.field", "protocol_version"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.value", "HTTP/1.0"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.field", "method"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.value", "HEAD"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.field", "domain"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.value", "example.com"),
				),
			},
			{
				Config: testAccNGWAFWorkspaceRuleConfigUpdate(workspaceName, updatedRuleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "description", updatedRuleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "type", "request"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "enabled", "false"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "request_logging", "none"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_operator", "any"),

					// Action
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "action.0.type", "allow"),

					// Conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.0.value", "10.0.0.1"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.1.value", "/admin"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.field", "agent_name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.operator", "matches"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "condition.2.value", "bot-*"),

					// Group conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.field", "country"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.value", "US"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.field", "method"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.1.value", "PUT"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.field", "protocol_version"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.value", "HTTP/2.0"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.field", "method"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.1.value", "OPTIONS"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.field", "domain"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.operator", "does_not_equal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.2.value", "internal.example"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace_rule.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rule := s.RootModule().Resources["fastly_ngwaf_workspace_rule.example"]
					workspace := s.RootModule().Resources["fastly_ngwaf_workspace.example"]
					return fmt.Sprintf("%s/%s", workspace.Primary.ID, rule.Primary.ID), nil
				},
			},
		},
	})
}

func testAccNGWAFWorkspaceRuleConfig(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_workspace_rule" "example" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "request"
  description      = "%s"
  enabled          = true
  request_logging  = "sampled"
  group_operator   = "all"

  action {
    type = "block"
  }

  condition {
    field    = "ip"
    operator = "equals"
    value    = "127.0.0.1"
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/login"
  }

  condition {
    field    = "agent_name"
    operator = "equals"
    value    = "host-001"
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "country"
      operator = "equals"
      value    = "AD"
    }

    condition {
      field    = "method"
      operator = "equals"
      value    = "POST"
    }
  }

  group_condition {
    group_operator = "any"

    condition {
      field    = "protocol_version"
      operator = "equals"
      value    = "HTTP/1.0"
    }

    condition {
      field    = "method"
      operator = "equals"
      value    = "HEAD"
    }

    condition {
      field    = "domain"
      operator = "equals"
      value    = "example.com"
    }
  }
}
`, workspaceName, ruleName)
}

func testAccNGWAFWorkspaceRuleConfigUpdate(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {}
}

resource "fastly_ngwaf_workspace_rule" "example" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "request"
  description      = "%s"
  enabled          = false
  request_logging  = "none"
  group_operator   = "any"

  action {
    type = "allow"
  }

  condition {
    field    = "ip"
    operator = "does_not_equal"
    value    = "10.0.0.1"
  }

  condition {
    field    = "path"
    operator = "does_not_equal"
    value    = "/admin"
  }

  condition {
    field    = "agent_name"
    operator = "matches"
    value    = "bot-*"
  }

  group_condition {
    group_operator = "any"

    condition {
      field    = "country"
      operator = "does_not_equal"
      value    = "US"
    }

    condition {
      field    = "method"
      operator = "does_not_equal"
      value    = "PUT"
    }
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "protocol_version"
      operator = "does_not_equal"
      value    = "HTTP/2.0"
    }

    condition {
      field    = "method"
      operator = "does_not_equal"
      value    = "OPTIONS"
    }

    condition {
      field    = "domain"
      operator = "does_not_equal"
      value    = "internal.example"
    }
  }
}
`, workspaceName, ruleName)
}
