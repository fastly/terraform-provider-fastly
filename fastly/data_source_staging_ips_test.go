package fastly

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceStagingIPs_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceStagingIPsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_staging_ips.example"]
						a := r.Primary.Attributes

						domains, err := strconv.Atoi(a["domains.#"])
						if err != nil {
							return err
						}

						if domains < 1 {
							return fmt.Errorf("expected at least one domain with a staging IP, got %d", domains)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceStagingIPsConfig(h string) string {
	tf := `
resource "fastly_service_vcl" "example" {
  name = "tf_example_service_for_staging_ips_data_source"

  domain {
    name = "%s.com"
  }

  backend {
    address = "httpbin.org"
    name    = "tf-test backend"
  }

  force_destroy = true
}

data "fastly_staging_ips" "example" {
  depends_on      = [fastly_service_vcl.example]
  service_id      = fastly_service_vcl.example.id
  service_version = fastly_service_vcl.example.active_version
}
`

	return fmt.Sprintf(tf, h)
}
