package fastly

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyDataSourceTLSConfigurationIDs(t *testing.T) {
	// lintignore:XAT001
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
