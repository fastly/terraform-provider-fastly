package servicecomputeauto_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/terraform-provider-fastly/internal/provider"
	testfixtures "github.com/fastly/terraform-provider-fastly/internal/test_fixtures"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestAccFastlyServiceComputeAuto_basic(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckServiceComputeAutoDestroy,
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "force_destroy", "true"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "active_version"),
					resource.TestCheckResourceAttrSet("fastly_service_compute_auto.test", "managed_version"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withBackend(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckServiceComputeAutoDestroy,
		Steps: []resource.TestStep{
			{
				Config: configWithBackend(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.port", "443"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.use_ssl", "true"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_update(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	serviceNameUpdated := fmt.Sprintf("tf-test-updated-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckServiceComputeAutoDestroy,
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
				),
			},
			{
				Config: configBasic(serviceNameUpdated, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_multipleBackends(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckServiceComputeAutoDestroy,
		Steps: []resource.TestStep{
			{
				Config: configMultipleBackends(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.0.name", "backend-primary"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "backend.1.name", "backend-secondary"),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_import(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccCheckServiceComputeAutoDestroy,
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceComputeAutoExists("fastly_service_compute_auto.test"),
				),
			},
			{
				ResourceName:            "fastly_service_compute_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "package", "reuse"},
			},
		},
	})
}

// Helper functions

func testAccPreCheck(t *testing.T) {
	if os.Getenv("FASTLY_API_TOKEN") == "" {
		t.Fatal("FASTLY_API_TOKEN must be set for acceptance tests")
	}
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"fastly": providerserver.NewProtocol6WithError(provider.New()),
	}
}

func testAccCheckServiceComputeAutoDestroy(s *terraform.State) error {
	apiToken := os.Getenv("FASTLY_API_TOKEN")
	client, err := fastly.NewClient(apiToken)
	if err != nil {
		return fmt.Errorf("error creating Fastly client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute_auto" {
			continue
		}

		service, err := client.GetServiceDetails(context.Background(), &fastly.GetServiceDetailsInput{
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

		return fmt.Errorf("Service %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckServiceComputeAutoExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		apiToken := os.Getenv("FASTLY_API_TOKEN")
		client, err := fastly.NewClient(apiToken)
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		_, err = client.GetServiceDetails(context.Background(), &fastly.GetServiceDetailsInput{
			ServiceID: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error fetching service: %w", err)
		}

		return nil
	}
}

// Configuration templates

func getPackagePath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	return filepath.Join(wd, "test_fixtures", "package", "valid.tar.gz")
}

func configBasic(serviceName, domainName string) string {
	return testfixtures.BuildConfig(
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": getPackagePath(),
		},
		"internal/resources/servicecomputeauto/test_fixtures/configs/basic.tf",
		"internal/test_fixtures/blocks/domain_single.tf",
		"internal/resources/servicecomputeauto/test_fixtures/blocks/package.tf",
	)
}

func configWithBackend(serviceName, domainName, backendName string) string {
	return testfixtures.BuildConfig(
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"BACKEND_NAME": backendName,
			"PACKAGE_PATH": getPackagePath(),
		},
		"internal/resources/servicecomputeauto/test_fixtures/configs/basic.tf",
		"internal/test_fixtures/blocks/domain_single.tf",
		"internal/test_fixtures/blocks/backend_single.tf",
		"internal/resources/servicecomputeauto/test_fixtures/blocks/package.tf",
	)
}

func configMultipleBackends(serviceName, domainName string) string {
	return testfixtures.BuildConfig(
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": getPackagePath(),
		},
		"internal/resources/servicecomputeauto/test_fixtures/configs/basic.tf",
		"internal/test_fixtures/blocks/domain_single.tf",
		"internal/test_fixtures/blocks/backend_multiple.tf",
		"internal/resources/servicecomputeauto/test_fixtures/blocks/package.tf",
	)
}
