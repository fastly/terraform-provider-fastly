package acceptancetests

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyACLEntries_create(t *testing.T) {
	t.Parallel()
	PreCheckAcc(t)

	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclID := CreateTestACL(t, aclName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLEntriesDestroy(aclID),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntries(aclID, map[string]string{
					"192.0.2.0/24":    "ALLOW",
					"198.51.100.0/24": "BLOCK",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "entries.%", "2"),
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "entries.192.0.2.0/24", "ALLOW"),
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "entries.198.51.100.0/24", "BLOCK"),
					CheckStandaloneACLEntriesRemoteState("fastly_acl_entries.acl_entries", map[string]string{
						"192.0.2.0/24":    "ALLOW",
						"198.51.100.0/24": "BLOCK",
					}),
				),
			},
			{
				ResourceName:            "fastly_acl_entries.acl_entries",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyACLEntries_update(t *testing.T) {
	t.Parallel()
	PreCheckAcc(t)

	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclID := CreateTestACL(t, aclName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLEntriesDestroy(aclID),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntries(aclID, map[string]string{
					"192.0.2.0/24":    "ALLOW",
					"198.51.100.0/24": "BLOCK",
				}),
				Check: resource.ComposeTestCheckFunc(
					CheckStandaloneACLEntriesRemoteState("fastly_acl_entries.acl_entries", map[string]string{
						"192.0.2.0/24":    "ALLOW",
						"198.51.100.0/24": "BLOCK",
					}),
				),
			},
			{
				Config: ConfigACLEntries(aclID, map[string]string{
					"203.0.113.0/24":  "BLOCK",
					"198.51.100.0/24": "ALLOW",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "entries.%", "2"),
					CheckStandaloneACLEntriesRemoteState("fastly_acl_entries.acl_entries", map[string]string{
						"203.0.113.0/24":  "BLOCK",
						"198.51.100.0/24": "ALLOW",
					}),
				),
			},
		},
	})
}

func TestAccFastlyACLEntries_delete(t *testing.T) {
	t.Parallel()
	PreCheckAcc(t)

	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclID := CreateTestACL(t, aclName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLEntriesDestroy(aclID),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntries(aclID, map[string]string{
					"192.0.2.0/24": "ALLOW",
				}),
				Check: resource.ComposeTestCheckFunc(
					CheckStandaloneACLEntriesRemoteState("fastly_acl_entries.acl_entries", map[string]string{
						"192.0.2.0/24": "ALLOW",
					}),
				),
			},
			{
				Config: ConfigACLEntries(aclID, map[string]string{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "entries.%", "0"),
					CheckStandaloneACLEntriesRemoteState("fastly_acl_entries.acl_entries", map[string]string{}),
				),
			},
		},
	})
}

// TestAccFastlyACLEntries_manageEntriesFalseSuppressDrift verifies
// that when manage_entries = false, changing the entries map in config
// produces no plan diff -- the provider must not apply or show the change.
func TestAccFastlyACLEntries_manageEntriesFalseSuppressDrift(t *testing.T) {
	t.Parallel()
	PreCheckAcc(t)

	aclName := fmt.Sprintf("tf_test_acl_%s", acctest.RandString(10))
	aclID := CreateTestACL(t, aclName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckACLEntriesDestroy(aclID),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesUnmanaged(aclID, map[string]string{
					"192.0.2.0/24": "ALLOW",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_acl_entries.acl_entries", "manage_entries", "false"),
				),
			},
			{
				Config: ConfigACLEntriesUnmanaged(aclID, map[string]string{
					"203.0.113.0/24": "BLOCK",
				}),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccFastlyACLEntries_invalidPrefix(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntries("fake-acl-id", map[string]string{
					"not_a_cidr": "ALLOW",
				}),
				ExpectError: regexp.MustCompile("not a valid CIDR prefix"),
			},
		},
	})
}

func TestAccFastlyACLEntries_invalidAction(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntries("fake-acl-id", map[string]string{
					"192.0.2.0/24": "PERMIT",
				}),
				ExpectError: regexp.MustCompile("must be either ALLOW or BLOCK"),
			},
		},
	})
}

// CheckACLEntriesDestroy returns a TestCheckFunc verifying that Terraform destroy
// cleared every entry from the given ACL. The ACL itself is a fixture created
// out-of-band via CreateTestACL, so its own lifecycle isn't asserted here.
func CheckACLEntriesDestroy(aclID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		resp, err := computeacls.ListEntries(context.Background(), client, &computeacls.ListEntriesInput{
			ComputeACLID: &aclID,
		})
		if err != nil {
			return fmt.Errorf("error listing ACL entries: %w", err)
		}

		if len(resp.Entries) != 0 {
			return fmt.Errorf("expected ACL %s to have no entries after destroy, found %d", aclID, len(resp.Entries))
		}

		return nil
	}
}

func CheckStandaloneACLEntriesRemoteState(resourceName string, want map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		id := rs.Primary.Attributes["acl_id"]
		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		resp, err := computeacls.ListEntries(context.Background(), client, &computeacls.ListEntriesInput{
			ComputeACLID: &id,
		})
		if err != nil {
			return fmt.Errorf("error listing ACL entries: %w", err)
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

func ConfigACLEntries(aclID string, entries map[string]string) string {
	return fmt.Sprintf(`
resource "fastly_acl_entries" "acl_entries" {
  acl_id = %q
  entries        = %s
  manage_entries = true
}
`, aclID, entriesHCL(entries))
}

func ConfigACLEntriesUnmanaged(aclID string, entries map[string]string) string {
	return fmt.Sprintf(`
resource "fastly_acl_entries" "acl_entries" {
  acl_id = %q
  entries        = %s
}
`, aclID, entriesHCL(entries))
}

func entriesHCL(entries map[string]string) string {
	hcl := "{\n"
	for prefix, action := range entries {
		hcl += fmt.Sprintf("    %q = %q\n", prefix, action)
	}
	hcl += "  }"
	return hcl
}
