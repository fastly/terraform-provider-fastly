package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceCDNACLEntries_create(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesCreate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "manage_entries", "true"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_acl_entries.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manage_entries"},
			},
		},
	})
}

func TestAccFastlyServiceCDNACLEntries_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesCreate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				Config: ConfigACLEntriesUpdate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "2"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNACLEntries_delete(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesCreate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				Config: ConfigACLEntriesDelete(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "0"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 0),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNACLEntries_manageEntriesFalse(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesManageEntriesFalse(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "manage_entries", "false"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNACLEntries_manyEntries(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	entryCount := 250

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesManyEntries(serviceName, domainName, aclName, entryCount),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", fmt.Sprintf("%d", entryCount)),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, entryCount),
				),
			},
		},
	})
}

// TestAccFastlyServiceACLEntries_omittedOptionalFields exercises an entry
// block that leaves negated, subnet, and comment unset. The API returns an
// explicit value for negated (false) even when it wasn't configured, so this
// guards against the provider reporting "inconsistent result after apply"
// for optional, non-computed entry attributes.
func TestAccFastlyServiceCDNACLEntries_omittedOptionalFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesMinimalEntry(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.0.negated", "false"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				// Re-plan with the same minimal config to confirm no drift
				// diff appears for the fields left unset.
				Config:   ConfigACLEntriesMinimalEntry(serviceName, domainName, aclName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNACLEntries_manageEntriesFalseSuppressDrift verifies
// that when manage_entries=false, changing the entry block in config produces
// no plan diff — the provider must not apply or show the change.
func TestAccFastlyServiceCDNACLEntries_manageEntriesFalseSuppressDrift(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesManageEntriesFalse(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "manage_entries", "false"),
				),
			},
			{
				// Change the entry IP in config — with manage_entries=false this must
				// produce an empty plan (no diff shown, no apply performed).
				Config:             ConfigACLEntriesManageEntriesFalseDifferentIP(serviceName, domainName, aclName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccFastlyServiceCDNACLEntries_sameIPDifferentSubnet verifies that two
// entries sharing the same IP but different subnets are created and read back
// correctly, exercising the flatten key-collision fix.
func TestAccFastlyServiceCDNACLEntries_sameIPDifferentSubnet(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesSameIPDifferentSubnet(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "2"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 2),
				),
			},
			{
				// Re-plan with identical config to confirm no spurious diff.
				Config:             ConfigACLEntriesSameIPDifferentSubnet(serviceName, domainName, aclName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TestAccFastlyServiceCDNACLEntries_modifyExistingEntry changes only the comment on an
// existing entry, keeping its ip/subnet fixed. This must apply as an in-place update of
// that entry, not a delete-and-recreate at the same ip/subnet -- Fastly's batch ACL
// entries API rejects a create that collides with an entry not yet deleted in the same
// batch, so a content-keyed (rather than identity-keyed) diff would fail here.
func TestAccFastlyServiceCDNACLEntries_modifyExistingEntry(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLEntriesCreate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.0.comment", "Test entry"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				Config: ConfigACLEntriesCommentChanged(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.0.ip", "127.0.0.1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl_entries.test", "entry.0.comment", "Test entry - updated"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
		},
	})
}

func CheckACLEntriesRemoteState(serviceResource, aclName string, expectedCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceResource]
		if !ok {
			return fmt.Errorf("not found: %s", serviceResource)
		}

		serviceID := rs.Primary.ID
		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		acls, err := client.ListACLs(context.Background(), &fastly.ListACLsInput{
			ServiceID:      serviceID,
			ServiceVersion: 1,
		})
		if err != nil {
			return fmt.Errorf("error listing ACLs: %w", err)
		}

		var aclID string
		for _, acl := range acls {
			if acl.Name != nil && *acl.Name == aclName {
				aclID = *acl.ACLID
				break
			}
		}

		if aclID == "" {
			return fmt.Errorf("ACL %s not found", aclName)
		}

		paginator := client.GetACLEntries(context.Background(), &fastly.GetACLEntriesInput{
			ServiceID: serviceID,
			ACLID:     aclID,
		})

		var entries []*fastly.ACLEntry
		for paginator.HasNext() {
			results, err := paginator.GetNext()
			if err != nil {
				return fmt.Errorf("error getting ACL entries: %w", err)
			}
			entries = append(entries, results...)
		}

		if len(entries) != expectedCount {
			return fmt.Errorf("expected %d entries, got %d", expectedCount, len(entries))
		}

		return nil
	}
}
