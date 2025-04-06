package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"
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

		aclID := rs.Primary.ID

		// Try to get the ACL
		_, err := computeacls.Describe(conn, &computeacls.DescribeInput{
			ComputeACLID: gofastly.ToPointer(aclID),
		})
		if err == nil {
			return fmt.Errorf("ACL still exists: %s", rs.Primary.ID)
		}

		// Check if the error is a 404
		if httpErr, ok := err.(*gofastly.HTTPError); ok && httpErr.StatusCode == 404 {
			continue
		}

		return err
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
		aclID := rs.Primary.ID

		// Get the ACL
		computeACL, err := computeacls.Describe(conn, &computeacls.DescribeInput{
			ComputeACLID: gofastly.ToPointer(aclID),
		})
		if err != nil {
			return err
		}

		// Convert to legacy ACL struct for test compatibility
		acl.ACLID = gofastly.ToPointer(computeACL.ComputeACLID)
		acl.Name = gofastly.ToPointer(computeACL.Name)

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
  name = "%s"
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
  force_destroy = true
}
`, serviceName, aclName)
}
