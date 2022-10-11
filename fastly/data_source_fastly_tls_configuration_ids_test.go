package fastly

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceTLSConfiguration_IDs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyAccDataSourceTLSConfigurationIDs,
				Check:  resource.TestCheckResourceAttrSet("data.fastly_tls_configuration_ids.subject", "ids.#"),
			},
		},
	})
}

const testAccFastlyAccDataSourceTLSConfigurationIDs = `data "fastly_tls_configuration_ids" "subject" {}`
