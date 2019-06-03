package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var flattenConditionTests = []struct {
	name     string
	in       []*gofastly.Condition
	expected []map[string]interface{}
}{
	{
		name: "basic flatten",
		in: []*gofastly.Condition{
			{
				Name: "some amz condition", Priority: 10,
				Type: "REQUEST", Statement: `req.url ~ "^/yolo/"`,
			},
		},
		expected: []map[string]interface{}{
			{
				"name":      "some amz condition",
				"priority":  10,
				"type":      "REQUEST",
				"statement": "req.url ~ \"^/yolo/\"",
			},
		},
	},
}

func TestResourceFastlyFlattenConditions(t *testing.T) {

	for _, tt := range flattenConditionTests {
		t.Run(tt.name, func(t *testing.T) {

			actual := flattenConditions(tt.in)

			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Error matching:\nexpected: %#v\ngot: %#v", tt.expected, actual)
			}
		})
	}
}

func TestAccFastlyServiceV1_conditional_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	con1 := gofastly.Condition{
		Name:      "some amz condition",
		Priority:  10,
		Type:      "REQUEST",
		Statement: `req.url ~ "^/yolo/"`,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1ConditionConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ConditionalAttributes(&service, name, []*gofastly.Condition{&con1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "condition.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1ConditionalAttributes(service *gofastly.ServiceDetail, name string, conditions []*gofastly.Condition) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
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
					c.Version = service.ActiveVersion.Number
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

func testAccServiceV1ConditionConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  condition {
    name = "some amz condition"
    type = "REQUEST"

    statement = "req.url ~ \"^/yolo/\""

    priority = 10
  }

  force_destroy = true
}`, name, domain)
}
