package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceAPISecurityOperationTags_Basic(t *testing.T) {
	serviceNameSuffix := acctest.RandString(10)
	tagName := "tf-tag-" + acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceAPISecurityOperationTagsConfig(serviceNameSuffix, tagName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operation_tags.example", "id"),
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operation_tags.example", "total"),
					resource.TestCheckResourceAttrSet("data.fastly_api_security_operation_tags.example", "tags.#"),

					resource.TestCheckTypeSetElemNestedAttrs("data.fastly_api_security_operation_tags.example", "tags.*", map[string]string{
						"name": tagName,
					}),
				),
			},
		},
	})
}

func testAccFastlyDataSourceAPISecurityOperationTagsConfig(serviceNameSuffix, tagName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "svc1" {
  name          = "test-svc-1-%s"
  force_destroy = true

  backend {
    address = "example.com"
    name    = "tf-test-backend-1"
  }
}

resource "fastly_api_security_operation_tag" "tag" {
  service_id  = fastly_service_vcl.svc1.id
  name        = "%s"
  description = "tag for ds tags test"
}

data "fastly_api_security_operation_tags" "example" {
  service_id = fastly_service_vcl.svc1.id

  depends_on = [
    fastly_api_security_operation_tag.tag
  ]
}
`, serviceNameSuffix, tagName)
}
