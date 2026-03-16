package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceAPISecurityOperations_Pagination(t *testing.T) {
	serviceNameSuffix := acctest.RandString(10)
	opDomain := fmt.Sprintf("api-test-%s.example.com", acctest.RandString(10))

	method := "GET"

	// Create two operations that match same filters (method+domain), different paths.
	opPath1 := "/tf-test-1"
	opPath2 := "/tf-test-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAPISecurityOperationsPaginationConfig(serviceNameSuffix, method, opDomain, opPath1, opPath2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operations.example", "id"),

					// Total should reflect both operations (across pages).
					resource.TestCheckResourceAttr("data.fastly_api_security_operations.example", "total", "2"),

					// meta.limit may vary; just ensure it's present when the API returns it.
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operations.example", "limit_returned"),

					// limit=1 is page size; provider should fetch both pages => 2 ops.
					resource.TestCheckResourceAttr("data.fastly_api_security_operations.example", "operations.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_api_security_operations.example", "operations.*", map[string]string{
						"method": method,
						"domain": opDomain,
						"path":   opPath1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_api_security_operations.example", "operations.*", map[string]string{
						"method": method,
						"domain": opDomain,
						"path":   opPath2,
					}),
				),
			},
		},
	})
}

func testAccFastlyDataSourceAPISecurityOperationsPaginationConfig(serviceNameSuffix, method, opDomain, opPath1, opPath2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_api_security_operation" "op1" {
  service_id  = fastly_service_vcl.svc1.id
  method      = "%s"
  domain      = "%s"
  path        = "%s"
  description = "ds ops pagination test - 1"
}

resource "fastly_api_security_operation" "op2" {
  service_id  = fastly_service_vcl.svc1.id
  method      = "%s"
  domain      = "%s"
  path        = "%s"
  description = "ds ops pagination test - 2"
}

data "fastly_api_security_operations" "example" {
  service_id = fastly_service_vcl.svc1.id

  # Filter to just our two operations
  method = ["%s"]
  domain = ["%s"]

  # Page size = 1, provider should auto-fetch all pages
  limit = 1

  depends_on = [
    fastly_api_security_operation.op1,
    fastly_api_security_operation.op2
  ]
}
`, serviceNameSuffix, method, opDomain, opPath1, method, opDomain, opPath2, method, opDomain)
}
