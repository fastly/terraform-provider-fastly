package fastly

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccFastlyDataSourceTLSConfiguration(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSConfiguration,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "name"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "tls_protocols.#"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "http_protocols.#"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "bulk"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "default"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "created_at"),
					resource.TestCheckResourceAttrSet("data.fastly_tls_configuration.subject", "updated_at"),
				),
			},
		},
	})
}

const testAccFastlyDataSourceTLSConfiguration = `
data "fastly_tls_configuration" "subject" {
  default = true
  tls_service = "CUSTOM"
}
`
