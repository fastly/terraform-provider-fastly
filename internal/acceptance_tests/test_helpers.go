package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/terraform-provider-fastly/internal/errors"
	"github.com/fastly/terraform-provider-fastly/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ProtoV6ProviderFactories returns the provider factories for acceptance tests
func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fastly": providerserver.NewProtocol6WithError(provider.New()),
	}
}

// PreCheck ensures the required environment variables are set for acceptance tests
func PreCheck(t *testing.T) {
	if v := os.Getenv("FASTLY_API_TOKEN"); v == "" {
		t.Fatal("FASTLY_API_TOKEN must be set for acceptance tests")
	}
}

// NewFastlyClient creates a new Fastly API client for testing
func NewFastlyClient() (*fastly.Client, error) {
	apiToken := os.Getenv("FASTLY_API_TOKEN")
	if apiToken == "" {
		return nil, fmt.Errorf("FASTLY_API_TOKEN environment variable must be set")
	}
	return fastly.NewClient(apiToken)
}

// serviceDestroyCheckAttempts and serviceDestroyCheckInterval bound the retry loop in
// CheckServiceDestroy, which tolerates the Fastly API's soft-delete taking a moment to become
// visible on a subsequent read - most noticeable when many acceptance tests run in parallel.
const (
	serviceDestroyCheckAttempts = 5
	serviceDestroyCheckInterval = 2 * time.Second
)

// CheckServiceDestroy returns a TestCheckFunc that verifies a service resource was destroyed
func CheckServiceDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			var lastErr error
			for attempt := 1; attempt <= serviceDestroyCheckAttempts; attempt++ {
				lastErr = checkServiceDestroyed(client, rs.Primary.ID)
				if lastErr == nil {
					break
				}
				if attempt < serviceDestroyCheckAttempts {
					time.Sleep(serviceDestroyCheckInterval)
				}
			}
			if lastErr != nil {
				return lastErr
			}
		}

		return nil
	}
}

func checkServiceDestroyed(client *fastly.Client, serviceID string) error {
	svc, err := client.GetService(context.Background(), &fastly.GetServiceInput{
		ServiceID: serviceID,
	})

	if errors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error checking if service was destroyed: %w", err)
	}

	// Service exists - check if it's soft-deleted
	if svc != nil && svc.DeletedAt != nil {
		return nil
	}

	return fmt.Errorf("service %s still exists", serviceID)
}

// CheckServiceExists returns a TestCheckFunc that verifies a service resource exists
func CheckServiceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no service ID is set")
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		_, err = client.GetService(context.Background(), &fastly.GetServiceInput{
			ServiceID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error fetching service: %w", err)
		}

		return nil
	}
}

// GetPackagePath returns the path to the valid.tar.gz test package
// Assumes tests are always run from the acceptance_tests package directory
func GetPackagePath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	return filepath.Join(wd, "fixtures", "packages", "valid.tar.gz")
}

// TODO: Replace this when implementing ACL entries
// AddACLEntry adds an ACL entry to the specified ACL. This is used as a test side-effect
// to populate ACLs for testing force_destroy behavior. Returns a TestCheckFunc.
func AddACLEntry(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		serviceID := rs.Primary.ID
		aclName := rs.Primary.Attributes["acl.0.name"]
		activeVersion := rs.Primary.Attributes["active_version"]

		if serviceID == "" || aclName == "" || activeVersion == "" {
			return fmt.Errorf("service_id, acl name, or active_version not set in state")
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		version := new(int)
		if _, err := fmt.Sscanf(activeVersion, "%d", version); err != nil {
			return fmt.Errorf("error parsing active_version: %w", err)
		}

		acl, err := client.GetACL(context.Background(), &fastly.GetACLInput{
			ServiceID:      serviceID,
			ServiceVersion: *version,
			Name:           aclName,
		})
		if err != nil {
			return fmt.Errorf("error fetching ACL %s: %w", aclName, err)
		}

		ip := "192.168.0.1"
		_, err = client.CreateACLEntry(context.Background(), &fastly.CreateACLEntryInput{
			ServiceID: serviceID,
			ACLID:     *acl.ACLID,
			IP:        &ip,
		})

		if err != nil {
			return fmt.Errorf("error adding entry to ACL %s on service %s: %w", *acl.ACLID, serviceID, err)
		}

		return nil
	}
}

// Configuration helpers for CDN Auto service

// ConfigCDNAutoBasic returns a basic CDN auto service config with a single domain
func ConfigCDNAutoBasic(serviceName, domainName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
	)
}

// ConfigCDNAutoWithBackend returns a CDN auto service config with a domain and backend
func ConfigCDNAutoWithBackend(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"BACKEND_NAME": backendName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
	)
}

// ConfigCDNAutoMultipleBackends returns a CDN auto service config with multiple backends
func ConfigCDNAutoMultipleBackends(serviceName, domainName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_multiple.tf",
	)
}

// ConfigCDNAutoUnsortedBackendAndDomainBlocks returns a CDN auto service config
// with backend and domain blocks declared in non-sorted order.
func ConfigCDNAutoUnsortedBackendAndDomainBlocks(serviceName, domainBName, domainAName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":  serviceName,
			"DOMAIN_B_NAME": domainBName,
			"DOMAIN_A_NAME": domainAName,
		},
		"internal/acceptance_tests/blocks/domain_multiple_unsorted.tf",
		"internal/acceptance_tests/blocks/backend_multiple_unsorted.tf",
	)
}

// ConfigCDNAutoReversedBackendAndDomainBlocks returns a CDN auto service config
// with the same backend and domain blocks as ConfigCDNAutoUnsortedBackendAndDomainBlocks,
// but declared in the reverse order.
func ConfigCDNAutoReversedBackendAndDomainBlocks(serviceName, domainBName, domainAName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":  serviceName,
			"DOMAIN_B_NAME": domainBName,
			"DOMAIN_A_NAME": domainAName,
		},
		"internal/acceptance_tests/blocks/domain_multiple_reversed.tf",
		"internal/acceptance_tests/blocks/backend_multiple_reversed.tf",
	)
}

// ConfigCDNAutoWithACL returns a CDN auto service config with a domain and ACL
func ConfigCDNAutoWithACL(serviceName, domainName, aclName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"ACL_NAME":     aclName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/acl_single.tf",
	)
}

// ConfigCDNAutoWithBackendAndACL returns a CDN auto service config with domain, backend, and ACL
func ConfigCDNAutoWithBackendAndACL(serviceName, domainName, backendName, aclName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"BACKEND_NAME": backendName,
			"ACL_NAME":     aclName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
		"internal/acceptance_tests/blocks/acl_single.tf",
	)
}

// ConfigCDNAutoWithMultipleACLs returns a CDN auto service config with multiple ACLs
func ConfigCDNAutoWithMultipleACLs(serviceName, domainName, aclName1, aclName2 string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"ACL_NAME_1":   aclName1,
			"ACL_NAME_2":   aclName2,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/acl_multi.tf",
	)
}

// ConfigCDNAutoWithACLForceDestroy returns a CDN auto service config with an ACL that has force_destroy enabled
func ConfigCDNAutoWithACLForceDestroy(serviceName, domainName, aclName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"ACL_NAME":     aclName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/acl_with_force_destroy.tf",
	)
}

// ConfigCDNAutoWithGzip returns a CDN auto service config with a domain and a gzip configuration
func ConfigCDNAutoWithGzip(serviceName, domainName, gzipName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"GZIP_NAME":    gzipName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/gzip_single.tf",
	)
}

// ConfigCDNAutoWithGzipEmptyLists returns a CDN auto service config with a gzip configuration
// that explicitly sets content_types and extensions to empty lists
func ConfigCDNAutoWithGzipEmptyLists(serviceName, domainName, gzipName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"GZIP_NAME":    gzipName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/gzip_empty_lists.tf",
	)
}

// ConfigCDNAutoWithGzipContentTypesRemoved returns a CDN auto service config with a gzip
// configuration whose content_types attribute has been removed, leaving extensions set
func ConfigCDNAutoWithGzipContentTypesRemoved(serviceName, domainName, gzipName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"GZIP_NAME":    gzipName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/gzip_content_types_removed.tf",
	)
}

// ConfigCDNAutoWithGzipAllRemoved returns a CDN auto service config with a gzip configuration
// whose content_types and extensions attributes have both been removed
func ConfigCDNAutoWithGzipAllRemoved(serviceName, domainName, gzipName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"GZIP_NAME":    gzipName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/gzip_all_removed.tf",
	)
}

// ConfigCDNAutoWithMultipleGzips returns a CDN auto service config with multiple gzip configurations
func ConfigCDNAutoWithMultipleGzips(serviceName, domainName, gzipName1, gzipName2 string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"GZIP_NAME_1":  gzipName1,
			"GZIP_NAME_2":  gzipName2,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/gzip_multi.tf",
	)
}

// ConfigACLForImport returns a test configuration for importing an ACL
func ConfigACLForImport(serviceName, domainName, aclName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"ACL_NAME":        aclName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
	)
}

// ConfigACLAtVersion returns a service/domain/ACL config pinned to the given version,
// for exercising in-place version changes on the explicit fastly_service_cdn_acl resource.
func ConfigACLAtVersion(serviceName, domainName, aclName string, version int) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": fmt.Sprintf("%d", version),
			"ACL_NAME":        aclName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
	)
}

// Configuration helpers for Compute Auto service

// ConfigComputeAutoBasic returns a basic Compute auto service config with a domain and package
func ConfigComputeAutoBasic(serviceName, domainName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoWithBackend returns a Compute auto service config with a domain, backend, and package
func ConfigComputeAutoWithBackend(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"BACKEND_NAME": backendName,
			"PACKAGE_PATH": GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoWithKVStoreResourceLink returns a Compute auto service config with a
// domain, package, and a resource_link block pointing at a Terraform-managed fastly_kvstore
// (declared as a sibling resource, referenced by ID rather than a literal string).
func ConfigComputeAutoWithKVStoreResourceLink(serviceName, domainName, storeName, linkName string) string {
	kvStoreConfig := fmt.Sprintf(`
resource "fastly_kvstore" "store" {
  name = %q
}

`, storeName)

	return kvStoreConfig + BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":            serviceName,
			"DOMAIN_NAME":             domainName,
			"PACKAGE_PATH":            GetPackagePath(),
			"RESOURCE_LINK_NAME":      linkName,
			"RESOURCE_LINK_TARGET_ID": "fastly_kvstore.store.id",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/resource_link_ref.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoWithStandaloneKVStore returns a Compute auto service config (domain and
// package, no resource_link) alongside a separately declared, unlinked fastly_kvstore.
//
// The Fastly API doesn't allow deleting a KV Store in the same request that unlinks it from a
// service, so tests that remove a resource_link and then delete its target KV Store need this as
// an intermediate step: unlink first and let that settle, then delete the KV Store in a later step.
func ConfigComputeAutoWithStandaloneKVStore(serviceName, domainName, storeName string) string {
	kvStoreConfig := fmt.Sprintf(`
resource "fastly_kvstore" "store" {
  name = %q
}

`, storeName)

	return kvStoreConfig + ConfigComputeAutoBasic(serviceName, domainName)
}

// ConfigComputeAutoWithKVStoreResourceLinkTarget returns a Compute auto service config
// declaring two fastly_kvstore resources (kv1 and kv2), with the resource_link pointing at
// whichever is named by targetLabel. Both KV Stores stay declared regardless of which is
// targeted, so retargeting exercises the reconcile delete-old/create-new pass without deleting
// either KV Store.
func ConfigComputeAutoWithKVStoreResourceLinkTarget(serviceName, domainName, storeName1, storeName2, linkName, targetLabel string) string {
	kvStoreConfig := fmt.Sprintf(`
resource "fastly_kvstore" "kv1" {
  name = %q
}

resource "fastly_kvstore" "kv2" {
  name = %q
}

`, storeName1, storeName2)

	return kvStoreConfig + BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":            serviceName,
			"DOMAIN_NAME":             domainName,
			"PACKAGE_PATH":            GetPackagePath(),
			"RESOURCE_LINK_NAME":      linkName,
			"RESOURCE_LINK_TARGET_ID": fmt.Sprintf("fastly_kvstore.%s.id", targetLabel),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/resource_link_ref.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigKVStore returns a minimal standalone fastly_kvstore config.
func ConfigKVStore(name string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "store" {
  name = %q
}
`, name)
}

// ConfigKVStoreWithLocation returns a standalone fastly_kvstore config with an explicit
// location, for exercising the location attribute's plan-time validation and its
// replace-on-change behavior.
func ConfigKVStoreWithLocation(name, location string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "store" {
  name     = %q
  location = %q
}
`, name, location)
}

// ConfigKVStoreForceDestroy returns a standalone fastly_kvstore config with force_destroy set,
// for exercising deletion of a KV Store that still contains entries.
func ConfigKVStoreForceDestroy(name string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "store" {
  name          = %q
  force_destroy = true
}
`, name)
}

// ConfigKVStoresDataSource returns a config declaring three fastly_kvstore resources alongside
// a fastly_kvstores data source that depends on all three.
func ConfigKVStoresDataSource(h string) string {
	return fmt.Sprintf(`
resource "fastly_kvstore" "store_1" {
  name = "tf_%s_1"
}

resource "fastly_kvstore" "store_2" {
  name = "tf_%s_2"
}

resource "fastly_kvstore" "store_3" {
  name = "tf_%s_3"
}

data "fastly_kvstores" "example" {
  depends_on = [
    fastly_kvstore.store_1,
    fastly_kvstore.store_2,
    fastly_kvstore.store_3,
  ]
}
`, h, h, h)
}

// ConfigComputeAutoWithACLResourceLink returns a Compute auto service config with a
// domain, package, and a resource_link block pointing at a Terraform-managed fastly_acl
// (declared as a sibling resource, referenced by ID rather than a literal string).
func ConfigComputeAutoWithACLResourceLink(serviceName, domainName, aclName, linkName string) string {
	aclConfig := fmt.Sprintf(`
resource "fastly_acl" "acl" {
  name = %q
}

`, aclName)

	return aclConfig + BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":            serviceName,
			"DOMAIN_NAME":             domainName,
			"PACKAGE_PATH":            GetPackagePath(),
			"RESOURCE_LINK_NAME":      linkName,
			"RESOURCE_LINK_TARGET_ID": "fastly_acl.acl.id",
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/resource_link_ref.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoWithStandaloneACL returns a Compute auto service config (domain and
// package, no resource_link) alongside a separately declared, unlinked fastly_acl.
//
// The Fastly API doesn't allow deleting an ACL in the same request that unlinks it from a
// service, so tests that remove a resource_link and then delete its target ACL need this as an
// intermediate step: unlink first and let that settle, then delete the ACL in a later step.
func ConfigComputeAutoWithStandaloneACL(serviceName, domainName, aclName string) string {
	aclConfig := fmt.Sprintf(`
resource "fastly_acl" "acl" {
  name = %q
}

`, aclName)

	return aclConfig + ConfigComputeAutoBasic(serviceName, domainName)
}

// ConfigComputeAutoWithACLResourceLinkTarget returns a Compute auto service config
// declaring two fastly_acl resources (acl1 and acl2), with the resource_link pointing at
// whichever is named by targetLabel. Both ACLs stay declared regardless of which is targeted, so
// retargeting exercises the reconcile delete-old/create-new pass without deleting either ACL.
func ConfigComputeAutoWithACLResourceLinkTarget(serviceName, domainName, aclName1, aclName2, linkName, targetLabel string) string {
	aclConfig := fmt.Sprintf(`
resource "fastly_acl" "acl1" {
  name = %q
}

resource "fastly_acl" "acl2" {
  name = %q
}

`, aclName1, aclName2)

	return aclConfig + BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":            serviceName,
			"DOMAIN_NAME":             domainName,
			"PACKAGE_PATH":            GetPackagePath(),
			"RESOURCE_LINK_NAME":      linkName,
			"RESOURCE_LINK_TARGET_ID": fmt.Sprintf("fastly_acl.%s.id", targetLabel),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/resource_link_ref.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoMultipleBackends returns a Compute auto service config with multiple backends and a package
func ConfigComputeAutoMultipleBackends(serviceName, domainName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_multiple.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoUnsortedBackendAndDomainBlocks returns a Compute auto service config
// with backend and domain blocks declared in non-sorted order.
func ConfigComputeAutoUnsortedBackendAndDomainBlocks(serviceName, domainBName, domainAName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":  serviceName,
			"DOMAIN_B_NAME": domainBName,
			"DOMAIN_A_NAME": domainAName,
			"PACKAGE_PATH":  GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_multiple_unsorted.tf",
		"internal/acceptance_tests/blocks/backend_multiple_unsorted.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoReversedBackendAndDomainBlocks returns a Compute auto service config
// with the same backend and domain blocks as ConfigComputeAutoUnsortedBackendAndDomainBlocks,
// but declared in the reverse order.
func ConfigComputeAutoReversedBackendAndDomainBlocks(serviceName, domainBName, domainAName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":  serviceName,
			"DOMAIN_B_NAME": domainBName,
			"DOMAIN_A_NAME": domainAName,
			"PACKAGE_PATH":  GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_multiple_reversed.tf",
		"internal/acceptance_tests/blocks/backend_multiple_reversed.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// Configuration helpers for CDN service (explicit version management)

// ConfigServiceCDNBasic returns a basic CDN service config without any nested resources
func ConfigServiceCDNBasic(serviceName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
		},
	)
}

// ConfigServiceCDNWithComment returns a CDN service config with a comment
func ConfigServiceCDNWithComment(serviceName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "Managed by Terraform",
		},
	)
}

// ConfigServiceCDNWithDomain returns a CDN service config with a domain resource
func ConfigServiceCDNWithDomain(serviceName, domainName string, version int) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": fmt.Sprintf("%d", version),
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
	)
}

// ConfigServiceCDNWithBackend returns a CDN service config with a domain and backend resource
func ConfigServiceCDNWithBackend(serviceName, domainName, backendName string, version int) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"BACKEND_NAME":    backendName,
			"SERVICE_VERSION": fmt.Sprintf("%d", version),
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
	)
}

// ConfigServiceCDNWithVersionClone returns a CDN service config with a version clone action
func ConfigServiceCDNWithVersionClone(serviceName, domainName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/action_version_clone.tf",
	)
}

// ConfigServiceCDNWithVersionActivate returns a CDN service config with a version activate action
func ConfigServiceCDNWithVersionActivate(serviceName, domainName string, version int) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": fmt.Sprintf("%d", version),
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/action_version_activate.tf",
	)
}

// ConfigServiceCDNWithCloneAndActivate returns a CDN service config with both clone and activate actions
func ConfigServiceCDNWithCloneAndActivate(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"BACKEND_NAME":    backendName,
			"SERVICE_VERSION": "1",
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/action_version_clone.tf",
		"internal/acceptance_tests/blocks/action_version_activate.tf",
	)
}

// Configuration helpers for Compute service (explicit version management)

// ConfigServiceComputeBasic returns a basic Compute service config without any nested resources
func ConfigServiceComputeBasic(serviceName string) string {
	return BuildConfig(
		ServiceCompute,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
		},
	)
}

// ConfigServiceComputeWithComment returns a Compute service config with a comment
func ConfigServiceComputeWithComment(serviceName string) string {
	return BuildConfig(
		ServiceCompute,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "Managed by Terraform",
		},
	)
}

// ConfigServiceComputeWithACLResourceLink returns an explicit Compute service config
// with a fastly_acl linked into it via fastly_service_resource_link.
func ConfigServiceComputeWithACLResourceLink(serviceName, aclName, linkName string) string {
	return BuildConfig(
		ServiceCompute,
		map[string]string{
			"SERVICE_NAME":       serviceName,
			"SERVICE_COMMENT":    "",
			"ACL_NAME":           aclName,
			"RESOURCE_LINK_NAME": linkName,
			"SERVICE_VERSION":    "1",
		},
		"internal/acceptance_tests/blocks/resource_link_acl.tf",
	)
}

// Configuration helpers for backend resources (explicit version management)

// ConfigBackendBasic returns a basic backend resource config
func ConfigBackendBasic(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_NAME":    backendName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_basic.tf",
	)
}

// ConfigBackendUpdated returns a backend resource config with updated values
func ConfigBackendUpdated(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_NAME":    backendName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_updated.tf",
	)
}

// ConfigBackendFull returns a backend resource config with all optional fields
func ConfigBackendFull(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_NAME":    backendName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_full.tf",
	)
}

func ConfigBackendFullUpdated(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_NAME":    backendName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_full_updated.tf",
	)
}

// ConfigBackendMultiple returns a config with multiple backend resources
func ConfigBackendMultiple(serviceName, domainName, backend1Name, backend2Name string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_1_NAME":  backend1Name,
			"BACKEND_2_NAME":  backend2Name,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_multi.tf",
	)
}

// ConfigBackendForImport returns a test configuration for importing a backend
func ConfigBackendForImport(serviceName, domainName, backendName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"BACKEND_NAME":    backendName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/backend_basic.tf",
	)
}

// Configuration helpers for domain resources (explicit version management)

// ConfigDomainBasic returns a basic domain resource config
func ConfigDomainBasic(serviceName, domainName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"SERVICE_VERSION": "1",
			"DOMAIN_NAME":     domainName,
		},
		"internal/acceptance_tests/blocks/domain_basic.tf",
	)
}

// ConfigDomainWithComment returns a domain resource config with a comment
func ConfigDomainWithComment(serviceName, domainName, comment string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"SERVICE_VERSION": "1",
			"DOMAIN_NAME":     domainName,
			"DOMAIN_COMMENT":  comment,
		},
		"internal/acceptance_tests/blocks/domain_with_comment.tf",
	)
}

// ConfigDomainMultiple returns a config with multiple domain resources
func ConfigDomainMultiple(serviceName, domain1Name, domain2Name string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"SERVICE_VERSION": "1",
			"DOMAIN_1_NAME":   domain1Name,
			"DOMAIN_2_NAME":   domain2Name,
		},
		"internal/acceptance_tests/blocks/domain_multi.tf",
	)
}

// ConfigDomainForImport returns a test configuration for importing a domain
func ConfigDomainForImport(serviceName, domainName, additionalDomainName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"SERVICE_VERSION": "1",
			"DOMAIN_1_NAME":   domainName,
			"DOMAIN_2_NAME":   additionalDomainName,
		},
		"internal/acceptance_tests/blocks/domain_multi.tf",
	)
}

// Configuration helpers for CDN service ACL entries resources

func aclEntriesBase(serviceName, domainName, aclName string) (ServiceType, map[string]string) {
	return ServiceCDN, map[string]string{
		"SERVICE_NAME":    serviceName,
		"SERVICE_COMMENT": "",
		"SERVICE_VERSION": "1",
		"DOMAIN_NAME":     domainName,
		"BACKEND_NAME":    "backend",
		"ACL_NAME":        aclName,
	}
}

func ConfigACLEntriesCreate(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_single.tf",
	)
}

func ConfigACLEntriesUpdate(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_two.tf",
	)
}

func ConfigACLEntriesDelete(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_empty.tf",
	)
}

func ConfigACLEntriesManageEntriesFalse(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_manage_false.tf",
	)
}

func ConfigACLEntriesManageEntriesFalseDifferentIP(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_manage_false_different_ip.tf",
	)
}

func ConfigACLEntriesMinimalEntry(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_minimal.tf",
	)
}

func ConfigACLEntriesSameIPDifferentSubnet(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_same_ip_different_subnet.tf",
	)
}

// ConfigACLEntriesCommentChanged mirrors ConfigACLEntriesCreate's single entry but with
// its comment changed and ip/subnet left untouched, exercising an in-place update of an
// existing entry rather than a replace or a create of an additional entry.
func ConfigACLEntriesCommentChanged(serviceName, domainName, aclName string) string {
	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_single_comment_changed.tf",
	)
}

func ConfigACLEntriesManyEntries(serviceName, domainName, aclName string, count int) string {
	var entries strings.Builder
	for i := 1; i <= count; i++ {
		fmt.Fprintf(&entries, "\n  entry {\n    ip      = \"%d.0.0.1\"\n    subnet  = 32\n    negated = false\n    comment = \"Entry %d\"\n  }", i, i)
	}
	entries.WriteString("\n")

	svc, replacements := aclEntriesBase(serviceName, domainName, aclName)
	replacements["ENTRIES"] = entries.String()
	return BuildConfig(svc, replacements,
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/service_cdn_backend.tf",
		"internal/acceptance_tests/blocks/acl_explicit.tf",
		"internal/acceptance_tests/blocks/acl_entries_many.tf",
	)
}

// Configuration helpers for the standalone Compute ACL entries resource (fastly_acl_entries)

// ConfigACLEntries returns a config declaring a fastly_acl resource alongside a
// fastly_acl_entries resource (with manage_entries = true) that targets it.
func ConfigACLEntries(aclName string, entries map[string]string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "acl" {
  name = %q
}

resource "fastly_acl_entries" "acl_entries" {
  acl_id         = fastly_acl.acl.id
  entries        = %s
  manage_entries = true
}
`, aclName, entriesHCL(entries))
}

// ConfigACLEntriesUnmanaged mirrors ConfigACLEntries but omits manage_entries,
// leaving it at its default (false).
func ConfigACLEntriesUnmanaged(aclName string, entries map[string]string) string {
	return fmt.Sprintf(`
resource "fastly_acl" "acl" {
  name = %q
}

resource "fastly_acl_entries" "acl_entries" {
  acl_id  = fastly_acl.acl.id
  entries = %s
}
`, aclName, entriesHCL(entries))
}

func entriesHCL(entries map[string]string) string {
	var hcl strings.Builder
	hcl.WriteString("{\n")
	for prefix, action := range entries {
		fmt.Fprintf(&hcl, "    %q = %q\n", prefix, action)
	}
	hcl.WriteString("  }")
	return hcl.String()
}

// Configuration helpers for S3 logging resources

// ConfigCDNAutoWithLoggingS3 returns a CDN auto service config with a domain and a nested S3 logging block.
func ConfigCDNAutoWithLoggingS3(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested.tf",
	)
}

// ConfigCDNAutoWithLoggingS3All returns a CDN auto service config with a nested S3 logging
// block that sets the full set of optional attributes.
func ConfigCDNAutoWithLoggingS3All(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested_all.tf",
	)
}

// ConfigCDNAutoWithLoggingS3GzipCodec returns a CDN auto service config with a nested S3 logging
// block that sets compression_codec = "gzip" and leaves gzip_level unset, exercising the auto
// read-back sentinel handling (MatchOrder/preserveGzipSentinelList) that must avoid a perpetual diff.
func ConfigCDNAutoWithLoggingS3GzipCodec(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested_gzip_codec.tf",
	)
}

// ConfigCDNAutoWithLoggingS3Updated returns a CDN auto service config with a nested S3 logging block
// whose optional attributes have been changed, exercising the reconcile update path.
func ConfigCDNAutoWithLoggingS3Updated(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested_updated.tf",
	)
}

// ConfigCDNAutoWithMultipleLoggingS3 returns a CDN auto service config with two nested S3 logging blocks.
func ConfigCDNAutoWithMultipleLoggingS3(serviceName, domainName, loggerName1, loggerName2, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":      serviceName,
			"DOMAIN_NAME":       domainName,
			"LOGGING_S3_NAME_1": loggerName1,
			"LOGGING_S3_NAME_2": loggerName2,
			"BUCKET_NAME":       bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested_multi.tf",
	)
}

// ConfigCDNAutoWithBackendAndLoggingS3 returns a CDN auto service config with a domain, backend, and nested S3 logging block.
func ConfigCDNAutoWithBackendAndLoggingS3(serviceName, domainName, backendName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDNAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"BACKEND_NAME":    backendName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/backend_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested.tf",
	)
}

// ConfigComputeAutoWithLoggingS3 returns a Compute auto service config with a domain, package, and nested S3 logging block.
func ConfigComputeAutoWithLoggingS3(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
			"PACKAGE_PATH":    GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

// ConfigComputeAutoWithLoggingS3Format returns a Compute auto service config whose
// nested S3 logging block sets format, a VCL-only attribute. service_compute_auto's
// logging_s3 schema (ComputeNestedBlockSchema) omits format/format_version/placement/
// response_condition entirely, so this is expected to fail Terraform's own schema
// validation ("Unsupported argument") rather than reach the Fastly API.
func ConfigComputeAutoWithLoggingS3Format(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"DOMAIN_NAME":     domainName,
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
			"PACKAGE_PATH":    GetPackagePath(),
		},
		"internal/acceptance_tests/blocks/domain_single.tf",
		"internal/acceptance_tests/blocks/logging_s3_nested_compute_format.tf",
		"internal/acceptance_tests/blocks/package.tf",
	)
}

func ConfigLoggingS3Basic(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_basic.tf",
	)
}

func ConfigLoggingS3AtVersion(serviceName, domainName, loggerName, bucketName string, version int) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": fmt.Sprintf("%d", version),
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_basic.tf",
	)
}

func ConfigLoggingS3NoAuth(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_no_auth.tf",
	)
}

func ConfigLoggingS3Updated(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_updated.tf",
	)
}

func ConfigLoggingS3IAM(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_iam.tf",
	)
}

func ConfigLoggingS3All(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_all.tf",
	)
}

func ConfigLoggingS3Defaults(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_defaults.tf",
	)
}

func ConfigLoggingS3CompressionCodec(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_compression_codec.tf",
	)
}

func ConfigLoggingS3GzipCodec(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_gzip_codec.tf",
	)
}

func ConfigLoggingS3CodecConflict(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_codec_conflict.tf",
	)
}

func ConfigLoggingS3ForImport(serviceName, domainName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCDN,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"DOMAIN_NAME":     domainName,
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/service_cdn_domain.tf",
		"internal/acceptance_tests/blocks/logging_s3_basic.tf",
	)
}

// ConfigLoggingS3ComputeFormat returns a config attaching fastly_service_logging_s3
// to an explicit Compute service with format set, a VCL-only attribute. Unlike the
// nested blocks, the standalone resource's schema is shared by both service types, so
// this is expected to fail at apply time via ValidateNoVCLOnlyAttributesForCompute
// rather than at Terraform's own schema-validation stage.
func ConfigLoggingS3ComputeFormat(serviceName, loggerName, bucketName string) string {
	return BuildConfig(
		ServiceCompute,
		map[string]string{
			"SERVICE_NAME":    serviceName,
			"SERVICE_COMMENT": "",
			"SERVICE_VERSION": "1",
			"LOGGING_S3_NAME": loggerName,
			"BUCKET_NAME":     bucketName,
		},
		"internal/acceptance_tests/blocks/logging_s3_compute_format.tf",
	)
}
