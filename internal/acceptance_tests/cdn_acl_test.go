package acceptancetests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceCDNAuto_withACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleACLs(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName1 := fmt.Sprintf("acl_1_%s", acctest.RandString(10))
	aclName2 := fmt.Sprintf("acl_2_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithMultipleACLs(serviceName, domainName, aclName1, aclName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName1),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.1.name", aclName2),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.1.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.1.force_destroy", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withBackendAndACL(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithBackendAndACL(serviceName, domainName, backendName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withACLUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "acl.0.acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withACLForceDestroy(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))
	aclNameUpdated := fmt.Sprintf("acl_updated_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "false"),
				),
			},
			{
				Config: ConfigCDNAutoWithACL(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					AddACLEntry("fastly_service_cdn_auto.test"),
				),
			},
			{
				Config:      ConfigCDNAutoWithACL(serviceName, domainName, aclNameUpdated),
				ExpectError: regexp.MustCompile("cannot delete ACL"),
			},
			{
				Config: ConfigCDNAutoWithACLForceDestroy(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "true"),
				),
			},
			{
				Config: ConfigCDNAutoWithACLForceDestroy(serviceName, domainName, aclNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.name", aclNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "acl.0.force_destroy", "true"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNACL_import(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLForImport(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "name", aclName),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_acl.test", "acl_id"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "version", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_cdn_acl.test"]
						if !ok {
							return fmt.Errorf("ACL resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_cdn_acl.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, aclName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNACL_versionUpdateInPlace verifies that changing only the
// version attribute on an explicit fastly_service_cdn_acl resource -- pointing it at
// a version that was cloned from the one it started on -- applies successfully
// and refreshes id/acl_id from the new version, rather than leaving them stale
// from the version the resource was created against.
func TestAccFastlyServiceCDNACL_versionUpdateInPlace(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s", acctest.RandString(10))

	var serviceID string
	var aclIDAtV1 string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLAtVersion(serviceName, domainName, aclName, 1),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_acl.test", "acl_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_cdn_acl.test"]
						if !ok {
							return fmt.Errorf("ACL resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						aclIDAtV1 = rs.Primary.Attributes["acl_id"]
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					client, err := NewFastlyClient()
					if err != nil {
						t.Fatalf("error creating Fastly client: %s", err)
					}
					if _, err := client.CloneVersion(context.Background(), &fastly.CloneVersionInput{
						ServiceID:      serviceID,
						ServiceVersion: 1,
					}); err != nil {
						t.Fatalf("error cloning version 1: %s", err)
					}
				},
				Config: ConfigACLAtVersion(serviceName, domainName, aclName, 2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "name", aclName),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "version", "2"),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_acl.test", "acl_id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_cdn_acl.test"]
						if !ok {
							return fmt.Errorf("ACL resource not found")
						}

						gotID := rs.Primary.Attributes["id"]
						wantID := fmt.Sprintf("%s-2-%s", serviceID, aclName)
						if gotID != wantID {
							return fmt.Errorf("expected id %q to reflect version 2, got %q", wantID, gotID)
						}

						client, err := NewFastlyClient()
						if err != nil {
							return fmt.Errorf("error creating Fastly client: %w", err)
						}
						remote, err := client.GetACL(context.Background(), &fastly.GetACLInput{
							ServiceID:      serviceID,
							ServiceVersion: 2,
							Name:           aclName,
						})
						if err != nil {
							return fmt.Errorf("error fetching ACL at version 2: %w", err)
						}

						gotACLID := rs.Primary.Attributes["acl_id"]
						if gotACLID != *remote.ACLID {
							return fmt.Errorf("state acl_id %q does not match version 2's ACL id %q (stale from version 1: %q)", gotACLID, *remote.ACLID, aclIDAtV1)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNACL_importWithUnderscores(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	aclName := fmt.Sprintf("acl_%s_with_underscores", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigACLForImport(serviceName, domainName, aclName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_acl.test", "name", aclName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_cdn_acl.test"]
						if !ok {
							return fmt.Errorf("ACL resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_cdn_acl.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, aclName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
