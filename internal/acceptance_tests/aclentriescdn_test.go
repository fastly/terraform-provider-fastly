package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceACLEntries_create(t *testing.T) {
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

func TestAccFastlyServiceACLEntries_update(t *testing.T) {
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
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", "1"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				Config: ConfigACLEntriesUpdate(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", "2"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceACLEntries_delete(t *testing.T) {
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
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", "1"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 1),
				),
			},
			{
				Config: ConfigACLEntriesDelete(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", "0"),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, 0),
				),
			},
		},
	})
}

func TestAccFastlyServiceACLEntries_manageEntriesFalse(t *testing.T) {
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
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "manage_entries", "false"),
				),
			},
		},
	})
}

func TestAccFastlyServiceACLEntries_manyEntries(t *testing.T) {
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
					resource.TestCheckResourceAttr("fastly_service_acl_entries.test", "entry.#", fmt.Sprintf("%d", entryCount)),
					CheckACLEntriesRemoteState("fastly_service_cdn.test", aclName, entryCount),
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

func ConfigACLEntriesCreate(serviceName, domainName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
}

resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  address    = "http-me.fastly.dev"
  name       = "backend"
}

resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_cdn_acl_entries" "test" {
  service_id      = fastly_service_cdn.test.id
  acl_id          = fastly_service_acl.test.acl_id
  manage_entries  = true

  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "Test entry"
  }
}
`, serviceName, domainName, aclName)
}

func ConfigACLEntriesUpdate(serviceName, domainName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
}

resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  address    = "http-me.fastly.dev"
  name       = "backend"
}

resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_cdn_acl_entries" "test" {
  service_id      = fastly_service_cdn.test.id
  acl_id          = fastly_service_acl.test.acl_id
  manage_entries  = true

  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "Test entry"
  }

  entry {
    ip      = "192.168.0.1"
    subnet  = "16"
    negated = true
    comment = "Second entry"
  }
}
`, serviceName, domainName, aclName)
}

func ConfigACLEntriesDelete(serviceName, domainName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
}

resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  address    = "http-me.fastly.dev"
  name       = "backend"
}

resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_cdn_acl_entries" "test" {
  service_id      = fastly_service_cdn.test.id
  acl_id          = fastly_service_acl.test.acl_id
  manage_entries  = true
}
`, serviceName, domainName, aclName)
}

func ConfigACLEntriesManageEntriesFalse(serviceName, domainName, aclName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
}

resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  address    = "http-me.fastly.dev"
  name       = "backend"
}

resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_cdn_acl_entries" "test" {
  service_id      = fastly_service_cdn.test.id
  acl_id          = fastly_service_acl.test.acl_id
  manage_entries  = false

  entry {
    ip      = "127.0.0.1"
    subnet  = "24"
    negated = false
    comment = "Test entry"
  }
}
`, serviceName, domainName, aclName)
}

func ConfigACLEntriesManyEntries(serviceName, domainName, aclName string, count int) string {
	entries := ""
	for i := 1; i <= count; i++ {
		entries += fmt.Sprintf(`
  entry {
    ip      = "%d.0.0.1"
    subnet  = "32"
    negated = false
    comment = "Entry %d"
  }
`, i, i)
	}

	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
}

resource "fastly_service_domain" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  address    = "http-me.fastly.dev"
  name       = "backend"
}

resource "fastly_service_acl" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_cdn_acl_entries" "test" {
  service_id      = fastly_service_cdn.test.id
  acl_id          = fastly_service_acl.test.acl_id
  manage_entries  = true
%s}
`, serviceName, domainName, aclName, entries)
}
