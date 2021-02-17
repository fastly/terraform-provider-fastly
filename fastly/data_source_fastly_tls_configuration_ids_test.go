package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccFastlyDataSourceTLSConfigurationIDs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyAccDataSourceTLSConfigurationIDs,
				Check:  resource.TestCheckResourceAttrSet("data.fastly_tls_configuration_ids.subject", "ids.#"),
			},
		},
	})
}

const testAccFastlyAccDataSourceTLSConfigurationIDs = `data "fastly_tls_configuration_ids" "subject" {}`
