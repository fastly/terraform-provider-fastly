package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceBackend_importBasic(t *testing.T) {
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

// ConfigBackendForImport returns a test configuration for importing a backend
func ConfigBackendForImport(serviceName, domainName, backendName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
  force_destroy = true
}

resource "fastly_service_domain" "test_domain" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_backend" "origin" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
  address    = "api.example.com"
  port       = 443
  use_ssl    = true
}
`, serviceName, domainName, backendName)
}
