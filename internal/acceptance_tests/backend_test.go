package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccFastlyServiceBackend_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendBasic(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "port", "443"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "use_ssl", "true"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_backend.origin", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_backend.origin", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceBackend_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendBasic(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "port", "443"),
				),
			},
			{
				Config: ConfigBackendUpdated(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.updated.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "port", "8443"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "comment", "Updated backend"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "weight", "50"),
				),
			},
		},
	})
}

func TestAccFastlyServiceBackend_fullConfig(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendFull(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "port", "443"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "use_ssl", "true"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "ssl_check_cert", "false"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "ssl_cert_hostname", "cert.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "ssl_sni_hostname", "sni.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "min_tls_version", "1.2"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "max_tls_version", "1.3"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "ssl_ciphers", "ECDHE-RSA-AES128-GCM-SHA256"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "weight", "75"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "max_conn", "100"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "connect_timeout", "2000"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "first_byte_timeout", "10000"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "between_bytes_timeout", "5000"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "error_threshold", "5"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "keepalive_time", "60"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "max_lifetime", "30000"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "max_use", "10"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "override_host", "override.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "shield", "iad-va-us"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "auto_loadbalance", "true"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "comment", "Full test backend"),
				),
			},
			{
				// keepalive_time is the only attribute without a Default (the API omits it when unset),
				// so it relies on UseStateForUnknown. This step omits it from config and changes an
				// unrelated field to verify it stays known in the plan rather than drifting to (known after apply).
				Config: ConfigBackendFullUpdated(serviceName, domainName, backendName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"fastly_service_backend.origin",
							tfjsonpath.New("keepalive_time"),
							knownvalue.Int64Exact(60),
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "weight", "50"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "keepalive_time", "60"),
				),
			},
		},
	})
}

func TestAccFastlyServiceBackend_multipleBackends(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backend1Name := fmt.Sprintf("backend-1-%s", acctest.RandString(10))
	backend2Name := fmt.Sprintf("backend-2-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendMultiple(serviceName, domainName, backend1Name, backend2Name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_backend.origin1", "name", backend1Name),
					resource.TestCheckResourceAttr("fastly_service_backend.origin1", "address", "api1.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin2", "name", backend2Name),
					resource.TestCheckResourceAttr("fastly_service_backend.origin2", "address", "api2.example.com"),
				),
			},
		},
	})
}

func TestAccFastlyServiceBackend_vclService(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-vcl-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendBasic(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
					CheckBackendExistsInFastly("fastly_service_cdn.test", backendName, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceBackend_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendForImport(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "port", "443"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "use_ssl", "true"),
					// Capture service_id and version for import step
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_backend.origin"]
						if !ok {
							return fmt.Errorf("backend resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import: service_id/version/name
				ResourceName: "fastly_service_backend.origin",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, backendName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyServiceBackend_importWithNameSlashes(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	// Backend name with slashes to test SplitN behavior
	backendName := fmt.Sprintf("backend/%s/with/slashes", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigBackendForImport(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_backend.origin", "name", backendName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_backend.origin"]
						if !ok {
							return fmt.Errorf("backend resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import with slashes in name
				ResourceName: "fastly_service_backend.origin",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, backendName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// CheckBackendExistsInFastly verifies a backend exists in Fastly API
func CheckBackendExistsInFastly(serviceName, backendName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		backend, err := client.GetBackend(context.Background(), &fastly.GetBackendInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           backendName,
		})
		if err != nil {
			return fmt.Errorf("error fetching backend from Fastly: %w", err)
		}

		if backend == nil {
			return fmt.Errorf("backend %s not found in Fastly", backendName)
		}

		return nil
	}
}
