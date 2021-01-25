package fastly

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccFastlyDataSourceTLSConfiguration_basic(t *testing.T) {
	resourceName := "data.fastly_tls_configuration.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSConfiguration_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "tls_protocols.#"),
					resource.TestCheckResourceAttrSet(resourceName, "http_protocols.#"),
					resource.TestCheckResourceAttr(resourceName, "tls_service", "CUSTOM"),
					resource.TestCheckResourceAttrSet(resourceName, "default"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

const testAccFastlyDataSourceTLSConfiguration_basic = `
data "fastly_tls_configuration" "subject" {
  default = true
  tls_service = "CUSTOM"
}
`

func TestAccFastlyDataSourceTLSConfiguration_withIDLookup(t *testing.T) {
	resourceName := "data.fastly_tls_configuration.subject"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSConfiguration_withIDLookup,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", "data.fastly_tls_configuration.default", "name"),
				),
			},
		},
	})
}

const testAccFastlyDataSourceTLSConfiguration_withIDLookup = `
data "fastly_tls_configuration" "default" {
  default = true
}
data "fastly_tls_configuration" "subject" {
  id = data.fastly_tls_configuration.default.id
}
`
