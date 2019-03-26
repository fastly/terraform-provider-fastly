package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func TestFastlyACLEntryV1(t *testing.T) {
	serviceName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))
	serviceDomain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	aclName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckACLEntryV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testACLEntryV1Config(serviceName, serviceDomain, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckACLEntryV1Exists("fastly_service_v1.foo", "fastly_acl_v1.foo", "fastly_acl_entry_v1.foo"),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "name", aclName),
					resource.TestCheckResourceAttr(
						"fastly_acl_v1.foo", "activate", "true"),
				),
			},
		},
	})
}

func testAccCheckACLEntryV1Exists(sn, an, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		serviceResource, ok := s.RootModule().Resources[sn]

		if !ok {
			return fmt.Errorf("Service not found: %s", sn)
		}

		if serviceResource.Primary.ID == "" {
			return fmt.Errorf("No service ID is set")
		}

		aclResource, ok := s.RootModule().Resources[an]

		if !ok {
			return fmt.Errorf("ACL not found: %s", an)
		}

		if aclResource.Primary.ID == "" {
			return fmt.Errorf("No ACL ID is set")
		}

		aclEntryResource, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("ACL entry not found: %s", n)
		}

		if aclEntryResource.Primary.ID == "" {
			return fmt.Errorf("No ACL entry ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		_, err := conn.GetACLEntry(&gofastly.GetACLEntryInput{
			Service: serviceResource.Primary.ID,
			ACL:     aclResource.Primary.ID,
			ID:      aclEntryResource.Primary.ID,
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckACLEntryV1Destroy(s *terraform.State) error {
	serviceResource, ok := s.RootModule().Resources["fastly_service_v1.foo"]

	if !ok {
		return fmt.Errorf("Service not found: %s", "fastly_service_v1.foo")
	}

	if serviceResource.Primary.ID == "" {
		return fmt.Errorf("No service ID is set")
	}

	aclResource, ok := s.RootModule().Resources["fastly_acl_v1.foo"]

	if !ok {
		return fmt.Errorf("ACL not found: %s", "fastly_acl_v1.foo")
	}

	aclEntryResource, ok := s.RootModule().Resources["fastly_acl_entry_v1.foo"]

	conn := testAccProvider.Meta().(*FastlyClient).conn
	entries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		Service: serviceResource.Primary.ID,
		ACL:     aclResource.Primary.ID,
	})

	if err != nil {
		return nil
	}

	for i := range entries {
		if entries[i].ID == aclEntryResource.Primary.ID {
			fmt.Errorf("[WARN] Tried deleting ACL entry (%s), but it was still found: %s", "fastly_acl_entry_v1.foo", err)
		}
	}

	return nil
}

func testACLEntryV1Config(serviceName, serviceDomain, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "amazon docs"
	}

	force_destroy = true
}

resource "fastly_acl_v1" "foo" {
  name       = "%s"
  service_id = "${fastly_service_v1.foo.id}"
}


resource "fastly_acl_entry_v1" "foo" {
  service_id = "${fastly_service_v1.foo.id}"
  acl_id     = "${fastly_acl_v1.foo.id}"
  ip         = "1.2.3.4"
  comment    = "terraform tests"
}
`, serviceName, serviceDomain, aclName)
}
