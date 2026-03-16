package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceAPISecurityDiscoveredOperations_Basic(t *testing.T) {
	serviceNameSuffix := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAPISecurityDiscoveredOperationsConfig(serviceNameSuffix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fastly_api_security_discovered_operations.example", "id"),
					resource.TestCheckResourceAttrSet("data.fastly_api_security_discovered_operations.example", "total"),

					// meta.limit may vary; just ensure it's present when the API returns it.
					resource.TestCheckResourceAttrSet("data.fastly_api_security_discovered_operations.example", "limit_returned"),

					resource.TestCheckResourceAttrSet("data.fastly_api_security_discovered_operations.example", "operations.#"),
				),
			},
		},
	})
}

// NOTE: Discovered operations depend on observed traffic and can be empty.
// This test mainly validates that the provider can call the endpoint and
// correctly reads pagination metadata.
func testAccFastlyDataSourceAPISecurityDiscoveredOperationsConfig(serviceNameSuffix string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

data "fastly_api_security_discovered_operations" "example" {
  service_id = fastly_service_vcl.svc1.id

  # Page size = 1 (provider will auto-fetch pages if there are any)
  limit = 1

  depends_on = [
    fastly_service_vcl.svc1
  ]
}
`, serviceNameSuffix)
}
