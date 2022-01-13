package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	var aclA gofastly.ACL
	var aclB gofastly.ACL
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.test", name)
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	// Six part test:
	// 1. Create service with 2 ACLs
	// 2. Rename both the ACLs, should succeed because the ACLs are empty
	// 3. Keep both ACLs the same and add an entry
	// 4. Try to rename the ACLs, expect to fail with "list not empty error"
	// 5. Without renaming the ACLs, set force_destroy=true to skip the deletion check
	// 6. Try to rename the ACLs again, expect to succeed
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_acl(name, aclName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclName, &aclA),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclName, &aclB),
				),
			},
			{
				Config: testAccServiceV1Config_acl(name, aclNameUpdated, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclNameUpdated, &aclA),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclNameUpdated, &aclB),
				),
			},
			{
				Config: testAccServiceV1Config_acl(name, aclNameUpdated, domain),
				// trigger side-effect of adding an ACL Entry
				Check: resource.ComposeTestCheckFunc(
					testAccAddACLEntries(&aclA),
					testAccAddACLEntries(&aclB),
				),
			},
			{
				Config:      testAccServiceV1Config_acl(name, aclName, domain),
				ExpectError: regexp.MustCompile("Cannot delete.*list is not empty.*"),
			},
			{
				Config: testAccServiceV1Config_aclForceDestroy(name, aclNameUpdated, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclNameUpdated, &aclA),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclNameUpdated, &aclB),
				),
			},
			{
				Config: testAccServiceV1Config_aclForceDestroy(name, aclName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclName, &aclA),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclName, &aclB),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_acl(service *gofastly.ServiceDetail, name, aclName string, acl *gofastly.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		remoteACL, err := conn.GetACL(&gofastly.GetACLInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
			Name:           aclName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up ACL records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if remoteACL.Name != aclName {
			return fmt.Errorf("ACL logging endpoint name mismatch, expected: %s, got: %#v", aclName, remoteACL.Name)
		}

		*acl = *remoteACL

		return nil
	}
}

// testAccAddACLEntries doesn't technically check for anything despite returning a TestCheckFunc. Instead it is used for
// its side effect of adding an ACL Entry
func testAccAddACLEntries(acl *gofastly.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		_, err := conn.CreateACLEntry(&gofastly.CreateACLEntryInput{
			ServiceID: acl.ServiceID,
			ACLID:     acl.ID,
			IP:        "192.168.0.1",
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error adding entry to ACL (%s) on service (%s): %w", acl.ID, acl.ServiceID, err)
		}

		return nil
	}
}

func testAccServiceV1Config_acl(name, aclName, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "127.0.0.1"
    name    = "tf-test-backend"
  }

  acl {
	name = "a_%s"
  }

  acl {
    name = "b_%s"
  }

  force_destroy = true
}`, name, domain, aclName, aclName)
}

func testAccServiceV1Config_aclForceDestroy(name, aclName, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "127.0.0.1"
    name    = "tf-test-backend"
  }

  acl {
	name          = "a_%s"
    force_destroy = true
  }

  acl {
    name          = "b_%s"
    force_destroy = true
  }

  force_destroy = true
}`, name, domain, aclName, aclName)
}
