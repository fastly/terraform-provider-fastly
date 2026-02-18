package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
)

func TestAccFastlyNGWAFAccountRule_basic(t *testing.T) {
	ruleDescription := fmt.Sprintf("Account Rule %s", acctest.RandString(5))
	updatedDescription := ruleDescription + " updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNGWAFAccountRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNGWAFAccountRuleConfig(ruleDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "description", ruleDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "enabled", "true"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "request_logging", "sampled"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "group_operator", "all"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "action.0.type", "block"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.field", "request_header"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.operator", "exists"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.group_operator", "any"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.0.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.0.operator", "contains"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.0.value", "Header-Sample"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.1.field", "name"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.1.operator", "equals"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "multival_condition.0.condition.1.value", "X-API-Key"),
				),
			},
			{
				Config: testAccNGWAFAccountRuleConfigUpdate(updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "description", updatedDescription),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "enabled", "false"),
					resource.TestCheckResourceAttr("fastly_ngwaf_account_rule.example", "action.0.type", "allow"),
				),
			},
			{
				ResourceName:      "fastly_ngwaf_account_rule.example",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rule := s.RootModule().Resources["fastly_ngwaf_account_rule.example"]
					return rule.Primary.ID, nil
				},
			},
		},
	})
}

func testAccNGWAFAccountRuleConfig(description string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_account_rule" "example" {
  applies_to       = ["*"]
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
    value    = "1.2.3.4"
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "method"
      operator = "equals"
      value    = "POST"
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
`, description)
}

func testAccNGWAFAccountRuleConfigUpdate(description string) string {
	return fmt.Sprintf(`
resource "fastly_ngwaf_account_rule" "example" {
  applies_to       = ["*"]
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

  group_condition {
    group_operator = "any"

    condition {
      field    = "method"
      operator = "equals"
      value    = "GET"
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
`, description)
}

func testAccCheckNGWAFAccountRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_ngwaf_account_rule" {
			continue
		}

		_, err := rules.Get(context.TODO(), conn, &rules.GetInput{
			RuleID: &rs.Primary.ID,
			Scope: &scope.Scope{
				Type:      scope.ScopeTypeAccount,
				AppliesTo: []string{"*"},
			},
		})
		if err == nil {
			return fmt.Errorf("NGWAF account rule %s still exists", rs.Primary.ID)
		}
	}

	return nil
}
