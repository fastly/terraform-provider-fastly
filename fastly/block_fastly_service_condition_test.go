package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenConditions(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Condition
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Condition{
				{
					Name:      gofastly.ToPointer("some amz condition"),
					Priority:  gofastly.ToPointer(10),
					Type:      gofastly.ToPointer("REQUEST"),
					Statement: gofastly.ToPointer(`req.url ~ "^/yolo/"`),
				},
			},
			local: []map[string]any{
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
		Comment:   gofastly.ToPointer(""),
		Name:      gofastly.ToPointer("some test condition"),
		Priority:  gofastly.ToPointer(10),
		Statement: gofastly.ToPointer(`req.url ~ "^/yolo/"`),
		Type:      gofastly.ToPointer("REQUEST"),
	}

	con2 := gofastly.Condition{
		Comment:   gofastly.ToPointer(""),
		Name:      gofastly.ToPointer("some test condition"),
		Priority:  gofastly.ToPointer(10),
		Statement: gofastly.ToPointer(`req.url ~ "^/yolo/"`),
		Type:      gofastly.ToPointer("CACHE"),
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
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLConditionalAttributes(&service, name, []*gofastly.Condition{&con1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "condition.#", "1"),
				),
			},
			{
				Config: testAccServiceVCLConditionConfigUpdate(name, domainName1, "CACHE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLConditionalAttributes(&service, name, []*gofastly.Condition{&con2}),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLConditionalAttributes(service *gofastly.ServiceDetail, name string, conditions []*gofastly.Condition) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		serviceName := gofastly.ToValue(service.Name)
		serviceVersionNumber := gofastly.ToValue(service.ActiveVersion.Number)

		if serviceName != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, serviceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		conditionList, err := conn.ListConditions(&gofastly.ListConditionsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: serviceVersionNumber,
		})
		if err != nil {
			return fmt.Errorf("error looking up Conditions for (%s), version (%v): %s", serviceName, serviceVersionNumber, err)
		}

		if len(conditionList) != len(conditions) {
			return fmt.Errorf("error: mismatch count of conditions, expected (%d), got (%d)", len(conditions), len(conditionList))
		}

		var found int
		for _, c := range conditions {
			for _, lc := range conditionList {
				if gofastly.ToValue(c.Name) == gofastly.ToValue(lc.Name) {
					// we don't know these things ahead of time, so populate them now
					c.ServiceID = service.ServiceID
					c.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					lc.CreatedAt = nil
					lc.UpdatedAt = nil
					if !reflect.DeepEqual(c, lc) {
						return fmt.Errorf("bad match Conditions match, expected (%#v), got (%#v)", c, lc)
					}
					found++
				}
			}
		}

		if found != len(conditions) {
			return fmt.Errorf("error matching Conditions rules")
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
