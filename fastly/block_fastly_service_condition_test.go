package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenConditions(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Condition
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Condition{
				{
					Name:      "some amz condition",
					Priority:  10,
					Type:      "REQUEST",
					Statement: `req.url ~ "^/yolo/"`,
				},
			},
			local: []map[string]interface{}{
				{
					"name":      "some amz condition",
					"priority":  10,
					"type":      "REQUEST",
					"statement": "req.url ~ \"^/yolo/\"",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenConditions(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_conditional_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	con1 := gofastly.Condition{
		Name:      "some test condition",
		Priority:  10,
		Type:      "REQUEST",
		Statement: `req.url ~ "^/yolo/"`,
	}

	con2 := gofastly.Condition{
		Name:      "some test condition",
		Priority:  10,
		Type:      "CACHE",
		Statement: `req.url ~ "^/yolo/"`,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConditionConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLConditionalAttributes(&service, name, []*gofastly.Condition{&con1}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "condition.#", "1"),
				),
			},
			{
				Config: testAccServiceVCLConditionConfigUpdate(name, domainName1, "CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLConditionalAttributes(&service, name, []*gofastly.Condition{&con2}),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLConditionalAttributes(service *gofastly.ServiceDetail, name string, conditions []*gofastly.Condition) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Conditions for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(conditionList) != len(conditions) {
			return fmt.Errorf("Error: mis match count of conditions, expected (%d), got (%d)", len(conditions), len(conditionList))
		}

		var found int
		for _, c := range conditions {
			for _, lc := range conditionList {
				if c.Name == lc.Name {
					// we don't know these things ahead of time, so populate them now
					c.ServiceID = service.ID
					c.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lc.CreatedAt = nil
					lc.UpdatedAt = nil
					if !reflect.DeepEqual(c, lc) {
						return fmt.Errorf("Bad match Conditions match, expected (%#v), got (%#v)", c, lc)
					}
					found++
				}
			}
		}

		if found != len(conditions) {
			return fmt.Errorf("Error matching Conditions rules")
		}
		return nil
	}
}

func testAccServiceVCLConditionConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  header {
    destination         = "http.x-foo"
    source              = "\"bar\""
    type                = "request"
    action              = "set"
    name                = "set x-foo"
    request_condition   = "some test condition"
  }

  condition {
    name = "some test condition"
    type = "REQUEST"

    statement = "req.url ~ \"^/yolo/\""

    priority = 10
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLConditionConfigUpdate(name, domain, condType string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  header {
    destination       = "http.x-foo"
    source            = "\"bar\""
    type              = "cache"
    action            = "set"
    name              = "set x-foo"
    cache_condition   = "some test condition"
  }

  condition {
    name = "some test condition"
    type = "%s"

    statement = "req.url ~ \"^/yolo/\""

    priority = 10
  }

  force_destroy = true
}`, name, domain, condType)
}
