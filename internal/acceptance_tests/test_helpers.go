package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
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

			svc, err := client.GetService(context.Background(), &fastly.GetServiceInput{
				ServiceID: rs.Primary.ID,
			})

			if errors.IsNotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error checking if service was destroyed: %w", err)
			}

			// Service exists - check if it's soft-deleted
			if svc != nil && svc.DeletedAt != nil {
				continue
			}

			return fmt.Errorf("service %s still exists", rs.Primary.ID)
		}

		return nil
	}
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
