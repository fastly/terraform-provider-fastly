package fastly

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/stretchr/testify/require"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
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
					Conditions: []rules.GroupConditionItem{
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "country", Operator: "equals", Value: "AD"}},
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "method", Operator: "equals", Value: "POST"}},
					},
				},
			},
			{
				Type: "group",
				Fields: rules.GroupCondition{
					GroupOperator: "any",
					Conditions: []rules.GroupConditionItem{
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "protocol_version", Operator: "equals", Value: "HTTP/1.0"}},
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "method", Operator: "equals", Value: "HEAD"}},
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "domain", Operator: "equals", Value: "example.com"}},
					},
				},
			},
			{
				Type: "group",
				Fields: rules.GroupCondition{
					GroupOperator: "all",
					Conditions: []rules.GroupConditionItem{
						{Type: "single", Fields: rules.Condition{Type: "single", Field: "ip", Operator: "in_list", Value: "site.mylist"}},
						{
							Type: "multival",
							Fields: rules.MultivalCondition{
								Field:         "request_header",
								Operator:      "exists",
								GroupOperator: "all",
								Conditions: []rules.ConditionMul{
									{Type: "single", Field: "name", Operator: "equals", Value: "X-myHeader"},
									{Type: "single", Field: "value_string", Operator: "equals", Value: "sampleString"},
								},
							},
						},
					},
				},
			},
			{
				Type: "multival",
				Fields: rules.MultivalCondition{
					Field:         "request_header",
					Operator:      "exists",
					GroupOperator: "any",
					Conditions: []rules.ConditionMul{
						{Field: "name", Operator: "contains", Value: "Header-Sample"},
						{Field: "name", Operator: "equals", Value: "X-API-Key"},
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
	require.Len(t, groups, 3)

	// Check the group with nested multival condition (third group)
	groupWithMultival := groups[2].(map[string]any)
	require.Equal(t, "all", groupWithMultival["group_operator"])

	// Check single conditions within the group
	groupConds := groupWithMultival["condition"].([]any)
	require.Len(t, groupConds, 1)
	require.Equal(t, "ip", groupConds[0].(map[string]any)["field"])
	require.Equal(t, "in_list", groupConds[0].(map[string]any)["operator"])
	require.Equal(t, "site.mylist", groupConds[0].(map[string]any)["value"])

	// Check nested multival condition within the group
	groupMultivals := groupWithMultival["multival_condition"].([]any)
	require.Len(t, groupMultivals, 1)

	nestedMultival := groupMultivals[0].(map[string]any)
	require.Equal(t, "request_header", nestedMultival["field"])
	require.Equal(t, "exists", nestedMultival["operator"])
	require.Equal(t, "all", nestedMultival["group_operator"])

	nestedMultivalConds := nestedMultival["condition"].([]any)
	require.Len(t, nestedMultivalConds, 2)
	require.Equal(t, "name", nestedMultivalConds[0].(map[string]any)["field"])
	require.Equal(t, "equals", nestedMultivalConds[0].(map[string]any)["operator"])
	require.Equal(t, "X-myHeader", nestedMultivalConds[0].(map[string]any)["value"])
	require.Equal(t, "value_string", nestedMultivalConds[1].(map[string]any)["field"])
	require.Equal(t, "equals", nestedMultivalConds[1].(map[string]any)["operator"])
	require.Equal(t, "sampleString", nestedMultivalConds[1].(map[string]any)["value"])

	// Multival conditions
	multivals := d.Get("multival_condition").([]any)
	require.Len(t, multivals, 1)

	multival := multivals[0].(map[string]any)
	require.Equal(t, "request_header", multival["field"])
	require.Equal(t, "exists", multival["operator"])
	require.Equal(t, "any", multival["group_operator"])

	multivalConds := multival["condition"].([]any)
	require.Len(t, multivalConds, 2)
	require.Equal(t, "name", multivalConds[0].(map[string]any)["field"])
	require.Equal(t, "contains", multivalConds[0].(map[string]any)["operator"])
	require.Equal(t, "Header-Sample", multivalConds[0].(map[string]any)["value"])
	require.Equal(t, "name", multivalConds[1].(map[string]any)["field"])
	require.Equal(t, "equals", multivalConds[1].(map[string]any)["operator"])
	require.Equal(t, "X-API-Key", multivalConds[1].(map[string]any)["value"])
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

					// Multival conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.operator", "contains"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.value", "Header-Sample"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.value", "X-API-Key"),
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

					// Multival conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.0.value", "Header-Sample-Updated"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.operator", "contains"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "multival_condition.0.condition.1.value", "X-API-Key-Updated"),
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

func TestAccFastlyNGWAFWorkspaceRule_nestedMultival(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	ruleDescription := fmt.Sprintf("Terraform Rule Nested Multival %s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil, // Rule is deleted implicitly when workspace is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceRuleConfigNestedMultival(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "type", "request"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "request_logging", "sampled"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_operator", "any"),

					// Action
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "action.0.type", "add_signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "action.0.signal", "site.test2"),

					// Group condition with nested single and multival conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.operator", "in_list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.condition.0.value", "site.mylist"),

					// Nested multival condition within group
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.0.value", "X-myHeader"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.1.field", "value_string"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.0.multival_condition.0.condition.1.value", "sampleString"),

					// Second group condition with different nested multival
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.operator", "in_list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.condition.0.value", "site.mylist"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.0.value", "X-myHeaderelse"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.1.field", "value_string"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.1.operator", "in_list"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.example", "group_condition.1.multival_condition.0.condition.1.value", "corp.test"),
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

  multival_condition {
    field          = "request_header"
    operator       = "exists"
    group_operator = "any"

    condition {
      field    = "name"
      operator = "contains"
      value    = "Header-Sample"
    }

    condition {
      field    = "name"
      operator = "equals"
      value    = "X-API-Key"
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

  multival_condition {
    field          = "request_header"
    operator       = "exists"
    group_operator = "any"

    condition {
      field    = "name"
      operator = "equals"
      value    = "Header-Sample-Updated"
    }

    condition {
      field    = "name"
      operator = "contains"
      value    = "X-API-Key-Updated"
    }
  }
}
`, workspaceName, ruleName)
}

func testAccNGWAFWorkspaceRuleConfigNestedMultival(workspaceName, ruleName string) string {
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
  group_operator   = "any"

  action {
    type   = "add_signal"
    signal = "site.test2"
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "ip"
      operator = "in_list"
      value    = "site.mylist"
    }

    multival_condition {
      field          = "request_header"
      operator       = "exists"
      group_operator = "all"

      condition {
        field    = "name"
        operator = "equals"
        value    = "X-myHeader"
      }

      condition {
        field    = "value_string"
        operator = "equals"
        value    = "sampleString"
      }
    }
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "ip"
      operator = "in_list"
      value    = "site.mylist"
    }

    multival_condition {
      field          = "request_header"
      operator       = "exists"
      group_operator = "any"

      condition {
        field    = "name"
        operator = "equals"
        value    = "X-myHeaderelse"
      }

      condition {
        field    = "value_string"
        operator = "in_list"
        value    = "corp.test"
      }
    }
  }
}
`, workspaceName, ruleName)
}

func TestAccFastlyNGWAFWorkspaceRule_rateLimit(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	ruleDescription := "some description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil, // Rule is deleted implicitly when workspace is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceRuleRateLimitConfig(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "type", "rate_limit"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "group_operator", "all"),

					// Conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.value", "/login"),

					// Rate limit
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.client_identifiers.0.type", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.duration", "500"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.interval", "60"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.signal", "site.test-signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.threshold", "100"),

					// Multival conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.operator", "contains"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.value", "Header-Sample"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.value", "X-API-Key"),
				),
			},
			{
				Config: testAccNGWAFWorkspaceRuleRateLimitConfigUpdate(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "type", "rate_limit"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "group_operator", "all"),

					// Conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "condition.0.value", "admin"),

					// Rate limit
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.client_identifiers.0.type", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.duration", "5000"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.interval", "600"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.signal", "site.test-signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "rate_limit.0.threshold", "1000"),

					// Multival conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.0.value", "Header-Sample-Updated"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.operator", "contains"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.rate_limit", "multival_condition.0.condition.1.value", "X-API-Key-Updated"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace_rule.rate_limit",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rule := s.RootModule().Resources["fastly_ngwaf_workspace_rule.rate_limit"]
					workspace := s.RootModule().Resources["fastly_ngwaf_workspace.example"]
					return fmt.Sprintf("%s/%s", workspace.Primary.ID, rule.Primary.ID), nil
				},
			},
		},
	})
}

func testAccNGWAFWorkspaceRuleRateLimitConfig(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_signal" "example" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  description      = "test signal"
  name             = "test signal"
}

resource "fastly_ngwaf_workspace_rule" "rate_limit" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "rate_limit"
  description      = "%s"
  enabled          = true
  group_operator   = "all"

  action {
    type   = "block_signal"
	signal = fastly_ngwaf_workspace_signal.example.reference_id
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/login"
  }

  rate_limit {
	client_identifiers {
		type = "ip"
	}
	duration  = 500
	interval  = 60
	signal    = fastly_ngwaf_workspace_signal.example.reference_id
	threshold = 100
  }

  multival_condition {
    field          = "request_header"
    operator       = "exists"
    group_operator = "any"

    condition {
      field    = "name"
      operator = "contains"
      value    = "Header-Sample"
    }

    condition {
      field    = "name"
      operator = "equals"
      value    = "X-API-Key"
    }
  }
}
`, workspaceName, ruleName)
}

func testAccNGWAFWorkspaceRuleRateLimitConfigUpdate(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_signal" "example" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  description      = "test signal"
  name             = "test signal"
}

resource "fastly_ngwaf_workspace_rule" "rate_limit" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "rate_limit"
  description      = "%s"
  enabled          = true
  group_operator   = "all"

  action {
    type   = "block_signal"
	signal = fastly_ngwaf_workspace_signal.example.reference_id
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "admin"
  }

  rate_limit {
	client_identifiers {
		type = "ip"
	}
	duration  = 5000
	interval  = 600
	signal    = fastly_ngwaf_workspace_signal.example.reference_id
	threshold = 1000
  }

  multival_condition {
    field          = "request_header"
    operator       = "exists"
    group_operator = "any"

    condition {
      field    = "name"
      operator = "equals"
      value    = "Header-Sample-Updated"
    }

    condition {
      field    = "name"
      operator = "contains"
      value    = "X-API-Key-Updated"
    }
  }
}
`, workspaceName, ruleName)
}

func TestAccFastlyNGWAFWorkspaceRule_templatedSignal(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	// the description must be an empty string for templated_signal rules
	ruleDescription := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil, // Rule is deleted implicitly when workspace is destroyed
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFWorkspaceRuleTemplatedSignalConfig(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "type", "templated_signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "group_operator", "all"),

					// Action for templated_signal
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "action.0.type", "templated_signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "action.0.signal", "2FA-CHANGED"),

					// Conditions
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.value", "127.0.0.1"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.value", "/login"),
				),
			},
			{
				Config: testAccNGWAFWorkspaceRuleTemplatedSignalConfigUpdate(workspaceName, ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace.example", "name", workspaceName),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "type", "templated_signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "group_operator", "all"),

					// Action should remain the same
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "action.0.type", "templated_signal"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "action.0.signal", "2FA-CHANGED"),

					// Condition should be updated (this tests ForceNew behavior)
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.field", "ip"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.0.value", "10.0.0.1"),

					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.field", "path"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_workspace_rule.templated_signal", "condition.1.value", "/admin"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_workspace_rule.templated_signal",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rule := s.RootModule().Resources["fastly_ngwaf_workspace_rule.templated_signal"]
					workspace := s.RootModule().Resources["fastly_ngwaf_workspace.example"]
					return fmt.Sprintf("%s/%s", workspace.Primary.ID, rule.Primary.ID), nil
				},
			},
		},
	})
}

func testAccNGWAFWorkspaceRuleTemplatedSignalConfig(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_rule" "templated_signal" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "templated_signal"
  description      = "%s"
  enabled          = true
  group_operator   = "all"

  action {
    type   = "templated_signal"
    signal = "2FA-CHANGED"
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
}
`, workspaceName, ruleName)
}

func testAccNGWAFWorkspaceRuleTemplatedSignalConfigUpdate(workspaceName, ruleName string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_workspace" "example" {
  name                            = "%s"
  description                     = "Test NGWAF Workspace"
  mode                            = "block"
  ip_anonymization                = "hashed"
  client_ip_headers               = ["X-Forwarded-For", "X-Real-IP"]
  default_blocking_response_code = 429

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}

resource "fastly_ngwaf_workspace_rule" "templated_signal" {
  workspace_id     = fastly_ngwaf_workspace.example.id
  type             = "templated_signal"
  description      = "%s"
  enabled          = true
  group_operator   = "all"

  action {
    type   = "templated_signal"
    signal = "2FA-CHANGED"
  }

  condition {
    field    = "ip"
    operator = "equals"
    value    = "10.0.0.1"
  }

  condition {
    field    = "path"
    operator = "equals"
    value    = "/admin"
  }
}
`, workspaceName, ruleName)
}

func TestAccFastlyNGWAFWorkspaceRule_emptyGroupCondition(t *testing.T) {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	ruleDescription := fmt.Sprintf("Terraform Rule %s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccNGWAFWorkspaceRuleConfigEmptyGroupCondition(workspaceName, ruleDescription),
				ExpectError: regexp.MustCompile(`group_condition\[0\]: must define at least one 'condition' or 'multival_condition' block`),
			},
		},
	})
}

func testAccNGWAFWorkspaceRuleConfigEmptyGroupCondition(workspaceName, ruleName string) string {
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

    group_condition {
      group_operator = "all"
      # Empty - no condition or multival_condition!
    }
  }
  `, workspaceName, ruleName)
}

func TestAccFastlyNGWAFWorkspaceRule_noConditions(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccFastlyNGWAFRuleConfigNoConditions(),
				ExpectError: regexp.MustCompile("rule must define at least one 'condition', 'group_condition', or 'multival_condition'"),
			},
		},
	})
}

func testAccFastlyNGWAFRuleConfigNoConditions() string {
	workspaceName := fmt.Sprintf("Test WAF Workspace %s", acctest.RandString(5))
	return fmt.Sprintf(`
  resource "fastly_ngwaf_workspace" "test" {
    name                            = "%s"
    description                     = "Test NGWAF Workspace"
    mode                            = "block"
    ip_anonymization                = "hashed"
    client_ip_headers               = ["X-Forwarded-For"]
    default_blocking_response_code = 429

    attack_signal_thresholds {}
  }

  resource "fastly_ngwaf_workspace_rule" "test" {
    workspace_id = fastly_ngwaf_workspace.test.id
    action {
      type = "block"
    }
    description = "Empty rule with no conditions"
    enabled = true
    type = "request"
  }
  `, workspaceName)
}
