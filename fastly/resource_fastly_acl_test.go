package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyACL_basic(t *testing.T) {
	var acl gofastly.ACL
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf-test-acl-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckFastlyACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyACLConfig(serviceName, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr("fastly_acl.test", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_acl.test", "acl_id"),
					resource.TestCheckResourceAttr("fastly_acl.test", "force_destroy", "false"),
				),
			},
			{
				Config: testAccFastlyACLConfigForceDestroy(serviceName, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr("fastly_acl.test", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_acl.test", "acl_id"),
					resource.TestCheckResourceAttr("fastly_acl.test", "force_destroy", "true"),
				),
			},
		},
	})
}

func TestAccFastlyACL_WithImport(t *testing.T) {
	var acl gofastly.ACL
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("tf-test-acl-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckFastlyACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyACLConfig(serviceName, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr("fastly_acl.test", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_acl.test", "acl_id"),
				),
			},
			{
				ResourceName:      "fastly_acl.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckFastlyACLDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_acl" {
			continue
		}

		serviceID, aclID, err := parseACLID(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Get the latest service version
		service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ServiceID: serviceID,
		})
		if err != nil {
			if httpErr, ok := err.(*gofastly.HTTPError); ok && httpErr.StatusCode == 404 {
				return nil
			}
			return err
		}

		var latestVersion int
		for _, version := range service.Versions {
			if version.Active != nil && *version.Active && version.Number != nil {
				if *version.Number > latestVersion {
					latestVersion = *version.Number
				}
			}
		}

		if latestVersion == 0 {
			// If no active version, use the latest version
			for _, version := range service.Versions {
				if version.Number != nil && *version.Number > latestVersion {
					latestVersion = *version.Number
				}
			}
		}

		acls, err := conn.ListACLs(&gofastly.ListACLsInput{
			ServiceID:      serviceID,
			ServiceVersion: latestVersion,
		})
		if err != nil {
			if httpErr, ok := err.(*gofastly.HTTPError); ok && httpErr.StatusCode == 404 {
				return nil
			}
			return err
		}

		for _, acl := range acls {
			if acl.ACLID != nil && *acl.ACLID == aclID {
				return fmt.Errorf("ACL still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckFastlyACLExists(name string, acl *gofastly.ACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ACL ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn

		serviceID, aclID, err := parseACLID(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Get the latest service version
		service, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ServiceID: serviceID,
		})
		if err != nil {
			return err
		}

		var latestVersion int
		for _, version := range service.Versions {
			if version.Active != nil && *version.Active && version.Number != nil {
				if *version.Number > latestVersion {
					latestVersion = *version.Number
				}
			}
		}

		if latestVersion == 0 {
			// If no active version, use the latest version
			for _, version := range service.Versions {
				if version.Number != nil && *version.Number > latestVersion {
					latestVersion = *version.Number
				}
			}
		}

		acls, err := conn.ListACLs(&gofastly.ListACLsInput{
			ServiceID:      serviceID,
			ServiceVersion: latestVersion,
		})
		if err != nil {
			return err
		}

		var found *gofastly.ACL
		for _, a := range acls {
			if a.ACLID != nil && *a.ACLID == aclID {
				found = a
				break
			}
		}

		if found == nil {
			return fmt.Errorf("ACL not found: %s", aclID)
		}

		*acl = *found
		log.Printf("[DEBUG] Found ACL: %s", aclID)

		return nil
	}
}

func testAccFastlyACLConfig(serviceName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "test" {
  name = "%s"

  domain {
    name    = "test.notadomain.com"
    comment = "not a domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}

resource "fastly_acl" "test" {
  name       = "%s"
  service_id = fastly_service_vcl.test.id
}
`, serviceName, aclName)
}

func testAccFastlyACLConfigForceDestroy(serviceName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "test" {
  name = "%s"

  domain {
    name    = "test.notadomain.com"
    comment = "not a domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}

resource "fastly_acl" "test" {
  name          = "%s"
  service_id    = fastly_service_vcl.test.id
  force_destroy = true
}
`, serviceName, aclName)
}
