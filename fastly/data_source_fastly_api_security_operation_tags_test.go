package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceAPISecurityOperationTags_Pagination(t *testing.T) {
	serviceNameSuffix := acctest.RandString(10)
	tagName1 := "tf-tag-" + acctest.RandString(8)
	tagName2 := "tf-tag-" + acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAPISecurityOperationTagsPaginationConfig(serviceNameSuffix, tagName1, tagName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operation_tags.example", "id"),
					resource.TestCheckResourceAttr("data.fastly_api_security_operation_tags.example", "total", "2"),
					resource.TestCheckResourceAttr("data.fastly_api_security_operation_tags.example", "limit_returned", "1"),

					// limit=1 is page size; provider should fetch both pages => 2 tags.
					resource.TestCheckResourceAttr("data.fastly_api_security_operation_tags.example", "tags.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_api_security_operation_tags.example", "tags.*", map[string]string{
						"name": tagName1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_api_security_operation_tags.example", "tags.*", map[string]string{
						"name": tagName2,
					}),
				),
			},
		},
	})
}

func testAccFastlyDataSourceAPISecurityOperationTagsPaginationConfig(serviceNameSuffix, tagName1, tagName2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_api_security_operation_tag" "tag1" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "%s"
  description = "tag for ds tags pagination test - 1"
}

resource "fastly_api_security_operation_tag" "tag2" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "%s"
  description = "tag for ds tags pagination test - 2"
}

data "fastly_api_security_operation_tags" "example" {
  service_id = fastly_service_vcl.svc1.id

  # Page size = 1, provider should auto-fetch all pages
  limit = 1

  depends_on = [
    fastly_api_security_operation_tag.tag1,
    fastly_api_security_operation_tag.tag2
  ]
}
`, serviceNameSuffix, tagName1, tagName2)
}
