package fastly

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/computeacls"
)

func TestResourceFastlyFlattenComputeACLEntries(t *testing.T) {
	cases := []struct {
		remote []computeacls.ComputeACLEntry
		local  map[string]string
	}{
		{
			remote: []computeacls.ComputeACLEntry{
				{
					Prefix: "192.0.2.0/24",
					Action: "ALLOW",
				},
				{
					Prefix: "198.51.100.0/24",
					Action: "BLOCK",
				},
			},
			local: map[string]string{
				"192.0.2.0/24":    "ALLOW",
				"198.51.100.0/24": "BLOCK",
			},
		},
	}

	for _, c := range cases {
		out := flattenComputeACLEntries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyComputeACLEntries_validate(t *testing.T) {
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))

	want1 := map[string]string{
		"192.0.2.0/24":    "ALLOW",
		"198.51.100.0/24": "BLOCK",
	}

	want2 := map[string]string{
		"203.0.113.0/24":  "BLOCK",
		"198.51.100.0/24": "ALLOW",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckFastlyComputeACLEntriesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeACLWithEntriesValidate(aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyComputeACLEntriesExist("fastly_compute_acl_entries.example"),
					testAccCheckFastlyComputeACLEntriesRemoteState("fastly_compute_acl_entries.example", want1),
				),
			},
			{
				Config: testAccComputeACLWithEntriesValidateUpdate(aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyComputeACLEntriesExist("fastly_compute_acl_entries.example"),
					testAccCheckFastlyComputeACLEntriesRemoteState("fastly_compute_acl_entries.example", want2),
				),
			},
			{
				ResourceName:            "fastly_compute_acl_entries.example",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyComputeACLEntries_invalidPrefix(t *testing.T) {
	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy: func(_ *terraform.State) error {
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeACLWithEntriesInvalid(aclName),
				ExpectError: regexp.MustCompile(".*expected valid CIDR notation.*"),
			},
		},
	})
}

func testAccCheckFastlyComputeACLEntriesExist(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		if r == nil {
			return fmt.Errorf("Not found: %s", n)
		}
		return nil
	}
}

func testAccCheckFastlyComputeACLEntriesRemoteState(n string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		if r == nil {
			return fmt.Errorf("Not found: %s", n)
		}

		id := r.Primary.Attributes["compute_acl_id"]
		conn := testAccProvider.Meta().(*APIClient).conn
		resp, err := computeacls.ListEntries(context.TODO(), conn, &computeacls.ListEntriesInput{
			ComputeACLID: &id,
		})
		if err != nil {
			return err
		}

		got := make(map[string]string)
		for _, entry := range resp.Entries {
			got[entry.Prefix] = entry.Action
		}

		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("error matching remote state:\nexpected: %#v\ngot: %#v", want, got)
		}
		return nil
	}
}

func testAccCheckFastlyComputeACLEntriesDestroy(s *terraform.State) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "fastly_compute_acl_entries" {
			continue
		}
		id := r.Primary.Attributes["compute_acl_id"]
		conn := testAccProvider.Meta().(*APIClient).conn
		resp, err := computeacls.ListEntries(context.TODO(), conn, &computeacls.ListEntriesInput{
			ComputeACLID: &id,
		})
		if err != nil {
			if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
				// Resource already gone — this is expected.
				return nil
			}
			return err
		}
		if len(resp.Entries) > 0 {
			return fmt.Errorf("expected no entries, got: %d", len(resp.Entries))
		}
	}
	return nil
}

func testAccComputeACLWithEntriesValidate(name string) string {
	return fmt.Sprintf(`
resource "fastly_compute_acl" "example" {
  name = "%s"
}

resource "fastly_compute_acl_entries" "example" {
  compute_acl_id  = fastly_compute_acl.example.id
  entries = {
    "192.0.2.0/24"    = "ALLOW"
    "198.51.100.0/24" = "BLOCK"
  }
  manage_entries = true
}
`, name)
}

func testAccComputeACLWithEntriesValidateUpdate(name string) string {
	return fmt.Sprintf(`
resource "fastly_compute_acl" "example" {
  name = "%s"
}

resource "fastly_compute_acl_entries" "example" {
  compute_acl_id  = fastly_compute_acl.example.id
  entries = {
    "203.0.113.0/24"  = "BLOCK"
    "198.51.100.0/24" = "ALLOW"
  }
  manage_entries = true
}
`, name)
}

func testAccComputeACLWithEntriesInvalid(name string) string {
	return fmt.Sprintf(`
resource "fastly_compute_acl" "example" {
  name = "%s"
}

resource "fastly_compute_acl_entries" "example" {
  compute_acl_id  = fastly_compute_acl.example.id
  entries = {
    "bad_cidr" = "ALLOW"
  }
  manage_entries = true
}
`, name)
}
