package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyACLEntries_basic(t *testing.T) {
	var acl gofastly.ACL
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	resourceName := "fastly_acl_entries.entries"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckFastlyACLEntriesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyACLEntriesConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "entry.0.prefix", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "entry.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "entry.1.prefix", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.1.action", "BLOCK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyACLEntries_update(t *testing.T) {
	var acl gofastly.ACL
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	resourceName := "fastly_acl_entries.entries"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckFastlyACLEntriesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyACLEntriesConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
				),
			},
			{
				Config: testAccFastlyACLEntriesConfigUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyACLExists("fastly_acl.test", &acl),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "entry.0.prefix", "192.168.0.0/24"),
					resource.TestCheckResourceAttr(resourceName, "entry.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "entry.1.prefix", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "entry.1.action", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "entry.2.prefix", "172.16.0.0/12"),
					resource.TestCheckResourceAttr(resourceName, "entry.2.action", "ALLOW"),
				),
			},
		},
	})
}

func testAccCheckFastlyACLEntriesDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*APIClient).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_acl_entries" {
			continue
		}

		aclID := rs.Primary.ID

		// Try to find the entries
		results, err := computeacls.ListEntries(conn, &computeacls.ListEntriesInput{
			ComputeACLID: gofastly.ToPointer(aclID),
		})
		if err == nil && results != nil && len(results.Entries) > 0 {
			return fmt.Errorf("ACL entries still exist: %s", aclID)
		}
	}

	return nil
}

func testAccFastlyACLEntriesConfig(name string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "test" {
  name = "%s"
  force_destroy = true
}

resource "fastly_acl_entries" "entries" {
  acl_id = fastly_acl.test.acl_id
  force_destroy = true

  entry {
    prefix = "192.168.0.0/24"
    action = "ALLOW"
  }

  entry {
    prefix = "10.0.0.0/8" 
    action = "BLOCK"
  }
}`, name)
}

func testAccFastlyACLEntriesConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "test" {
  name = "%s"
  force_destroy = true
}

resource "fastly_acl_entries" "entries" {
  acl_id = fastly_acl.test.acl_id
  force_destroy = true

  entry {
    prefix = "192.168.0.0/24"
    action = "ALLOW"
  }

  entry {
    prefix = "10.0.0.0/8" 
    action = "BLOCK"
  }

  entry {
    prefix = "172.16.0.0/12"
    action = "ALLOW"
  }
}`, name)
}