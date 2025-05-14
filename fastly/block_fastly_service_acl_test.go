package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

func TestResourceFastlyFlattenAcl(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ACL
		local  []map[string]any
	}{
		{
			remote: []*gofastly.ACL{
				{
					ACLID: gofastly.ToPointer("1234567890"),
					Name:  gofastly.ToPointer("acl-example"),
				},
			},
			local: []map[string]any{
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

func TestAccFastlyServiceVCL_acl(t *testing.T) {
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
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigACL(name, aclName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "a_"+aclName, &aclA),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "b_"+aclName, &aclB),
				),
			},
			{
				Config: testAccServiceVCLConfigACL(name, aclNameUpdated, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "a_"+aclNameUpdated, &aclA),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "b_"+aclNameUpdated, &aclB),
				),
			},
			{
				Config: testAccServiceVCLConfigACL(name, aclNameUpdated, domain),
				// trigger side-effect of adding an ACL Entry
				Check: resource.ComposeTestCheckFunc(
					testAccAddACLEntries(&aclA),
					testAccAddACLEntries(&aclB),
				),
			},
			{
				Config:      testAccServiceVCLConfigACL(name, aclName, domain),
				ExpectError: regexp.MustCompile("cannot delete.*list is not empty.*"),
			},
			{
				Config: testAccServiceVCLConfigACLForceDestroy(name, aclNameUpdated, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "a_"+aclNameUpdated, &aclA),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "b_"+aclNameUpdated, &aclB),
				),
			},
			{
				Config: testAccServiceVCLConfigACLForceDestroy(name, aclName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "a_"+aclName, &aclA),
					testAccCheckFastlyServiceVCLAttributesACL(&service, name, "b_"+aclName, &aclB),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLAttributesACL(service *gofastly.ServiceDetail, name, aclName string, acl *gofastly.ACL) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		serviceName := gofastly.ToValue(service.Name)

		if serviceName != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, serviceName)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		remoteACL, err := conn.GetACL(&gofastly.GetACLInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
			Name:           aclName,
		})
		if err != nil {
			return fmt.Errorf("error looking up ACL records for (%s), version (%v): %s", serviceName, gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if gofastly.ToValue(remoteACL.Name) != aclName {
			return fmt.Errorf("acl logging endpoint name mismatch, expected: %s, got: %#v", aclName, gofastly.ToValue(remoteACL.Name))
		}

		*acl = *remoteACL

		return nil
	}
}

// testAccAddACLEntries doesn't technically check for anything despite returning a TestCheckFunc. Instead it is used for
// its side effect of adding an ACL Entry.
func testAccAddACLEntries(acl *gofastly.ACL) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		_, err := conn.CreateACLEntry(&gofastly.CreateACLEntryInput{
			ServiceID: gofastly.ToValue(acl.ServiceID),
			ACLID:     gofastly.ToValue(acl.ACLID),
			IP:        gofastly.ToPointer("192.168.0.1"),
		})
		if err != nil {
			return fmt.Errorf("error adding entry to ACL (%s) on service (%s): %w", gofastly.ToValue(acl.ACLID), gofastly.ToValue(acl.ServiceID), err)
		}

		return nil
	}
}

func testAccServiceVCLConfigACL(name, aclName, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigACLForceDestroy(name, aclName, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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
