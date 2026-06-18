package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceDomain_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	additionalDomainName := fmt.Sprintf("www.%s.example.com", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainForImport(serviceName, domainName, additionalDomainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "name", additionalDomainName),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "comment", "Additional domain"),
					// Capture service_id and version for import step
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_domain.additional"]
						if !ok {
							return fmt.Errorf("domain resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import: service_id/version/name
				ResourceName: "fastly_service_domain.additional",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, additionalDomainName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFastlyServiceDomain_importWithSubdomain(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	// Domain with multiple levels to ensure proper parsing
	subdomainName := fmt.Sprintf("api.v2.staging.%s.example.com", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigDomainForImport(serviceName, domainName, subdomainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_domain.additional", "name", subdomainName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_domain.additional"]
						if !ok {
							return fmt.Errorf("domain resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				// Test legacy composite ID import with complex subdomain
				ResourceName: "fastly_service_domain.additional",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, subdomainName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// ConfigDomainForImport returns a test configuration for importing a domain
func ConfigDomainForImport(serviceName, domainName, additionalDomainName string) string {
	return fmt.Sprintf(`
resource "fastly_service_cdn" "test" {
  name = "%s"
  force_destroy = true
}

resource "fastly_service_domain" "primary" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
}

resource "fastly_service_domain" "additional" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "%s"
  comment    = "Additional domain"
}
`, serviceName, domainName, additionalDomainName)
}
