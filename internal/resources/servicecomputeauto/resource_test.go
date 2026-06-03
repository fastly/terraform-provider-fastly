package servicecomputeauto_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	acceptancetesthelpers "github.com/fastly/terraform-provider-fastly/internal/acceptance_test_helpers"
)

func TestAccFastlyServiceComputeAuto_basic(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptancetesthelpers.PreCheck(t) },
		ProtoV6ProviderFactories: acceptancetesthelpers.ProtoV6ProviderFactories(),
		CheckDestroy:             acceptancetesthelpers.CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
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
		PreCheck:                 func() { acceptancetesthelpers.PreCheck(t) },
		ProtoV6ProviderFactories: acceptancetesthelpers.ProtoV6ProviderFactories(),
		CheckDestroy:             acceptancetesthelpers.CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: configWithBackend(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
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
		PreCheck:                 func() { acceptancetesthelpers.PreCheck(t) },
		ProtoV6ProviderFactories: acceptancetesthelpers.ProtoV6ProviderFactories(),
		CheckDestroy:             acceptancetesthelpers.CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "domain.0.name", domainName),
				),
			},
			{
				Config: configBasic(serviceNameUpdated, domainName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
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
		PreCheck:                 func() { acceptancetesthelpers.PreCheck(t) },
		ProtoV6ProviderFactories: acceptancetesthelpers.ProtoV6ProviderFactories(),
		CheckDestroy:             acceptancetesthelpers.CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: configMultipleBackends(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
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
		PreCheck:                 func() { acceptancetesthelpers.PreCheck(t) },
		ProtoV6ProviderFactories: acceptancetesthelpers.ProtoV6ProviderFactories(),
		CheckDestroy:             acceptancetesthelpers.CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: configBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					acceptancetesthelpers.CheckServiceExists("fastly_service_compute_auto.test"),
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

// Configuration templates


func configBasic(serviceName, domainName string) string {
	return acceptancetesthelpers.BuildConfig(
		acceptancetesthelpers.ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": acceptancetesthelpers.GetPackagePath(),
		},
		"internal/acceptance_test_helpers/blocks/domain_single.tf",
		"internal/acceptance_test_helpers/blocks/package.tf",
	)
}

func configWithBackend(serviceName, domainName, backendName string) string {
	return acceptancetesthelpers.BuildConfig(
		acceptancetesthelpers.ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"BACKEND_NAME": backendName,
			"PACKAGE_PATH": acceptancetesthelpers.GetPackagePath(),
		},
		"internal/acceptance_test_helpers/blocks/domain_single.tf",
		"internal/acceptance_test_helpers/blocks/backend_single.tf",
		"internal/acceptance_test_helpers/blocks/package.tf",
	)
}

func configMultipleBackends(serviceName, domainName string) string {
	return acceptancetesthelpers.BuildConfig(
		acceptancetesthelpers.ServiceComputeAuto,
		map[string]string{
			"SERVICE_NAME": serviceName,
			"DOMAIN_NAME":  domainName,
			"PACKAGE_PATH": acceptancetesthelpers.GetPackagePath(),
		},
		"internal/acceptance_test_helpers/blocks/domain_single.tf",
		"internal/acceptance_test_helpers/blocks/backend_multiple.tf",
		"internal/acceptance_test_helpers/blocks/package.tf",
	)
}
