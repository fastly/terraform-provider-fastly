package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenAcl(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ACL
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.ACL{
				{
					ID:   "1234567890",
					Name: "acl-example",
				},
			},
			local: []map[string]interface{}{
				{
					"acl_id": "1234567890",
					"name":   "acl-example",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenACLs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_acl(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_acl(name, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclName),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclName),
				),
			},
			{
				Config: testAccServiceV1Config_acl(name, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclNameUpdated),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclNameUpdated),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_acl(service *gofastly.ServiceDetail, name, aclName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		acl, err := conn.GetACL(&gofastly.GetACLInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
			Name:           aclName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up ACL records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if acl.Name != aclName {
			return fmt.Errorf("ACL logging endpoint name mismatch, expected: %s, got: %#v", aclName, acl.Name)
		}

		return nil
	}
}

func testAccServiceV1Config_acl(name, aclName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf-test-backend"
  }

  acl {
	name = "a_%s"
  }

  acl {
    name = "b_%s"
  }

  force_destroy = true
}`, name, domainName, backendName, aclName, aclName)
}
