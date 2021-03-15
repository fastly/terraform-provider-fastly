package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
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
	var acl gofastly.ACL
	name := acctest.RandomWithPrefix(testResourcePrefix)
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	// Six part test:
	// 1. Create service with 2 ACLs
	// 2. Rename both the ACLs, should succeed because the ACLs are empty
	// 3. Keep both ACLs the same and add an entry to one of them
	// 4. Try to rename the ACLs, expect to fail with "list not empty error"
	// 5. Without renaming the ACLs, set force_destroy=true to skip the deletion check
	// 6. Try to rename the ACLs again, expect to succeed
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_acl(name, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclName, &acl),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclName, &acl),
				),
			},
			{
				Config: testAccServiceV1Config_acl(name, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclNameUpdated, &acl),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclNameUpdated, &acl),
				),
			},
			{
				Config: testAccServiceV1Config_acl(name, aclNameUpdated),
				Check:  testAccAddACLEntries(&acl), // triggers side-effect of adding an ACL Entry
			},
			{
				Config:      testAccServiceV1Config_acl(name, aclName),
				ExpectError: regexp.MustCompile("Cannot delete.*list is not empty.*"),
			},
			{
				Config: testAccServiceV1Config_aclForceDestroy(name, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclNameUpdated, &acl),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclNameUpdated, &acl),
				),
			},
			{
				Config: testAccServiceV1Config_aclForceDestroy(name, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "a_"+aclName, &acl),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, "b_"+aclName, &acl),
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

func testAccServiceV1Config_aclForceDestroy(name, aclName string) string {
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
	name          = "a_%s"
    force_destroy = true
  }

  acl {
    name          = "b_%s"
    force_destroy = true
  }

  force_destroy = true
}`, name, domainName, backendName, aclName, aclName)
}
