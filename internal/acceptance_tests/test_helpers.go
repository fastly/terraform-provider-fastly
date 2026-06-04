package acceptancetests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
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

			service, err := client.GetService(context.Background(), &fastly.GetServiceInput{
				ServiceID: rs.Primary.ID,
			})

			if fastlyErr, ok := err.(*fastly.HTTPError); ok && fastlyErr.StatusCode == 404 {
				continue
			}

			if err != nil {
				return fmt.Errorf("error checking if service was destroyed: %w", err)
			}

			// Service exists - check if it's soft-deleted
			if service != nil && service.DeletedAt != nil {
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
