package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceDNSZones_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceDNSZonesConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_dns_zones.example"]
						a := r.Primary.Attributes

						want := []string{
							fmt.Sprintf("tf-%s-1.fastly-example.com.", h),
							fmt.Sprintf("tf-%s-2.fastly-example.com.", h),
							fmt.Sprintf("tf-%s-3.fastly-example.com.", h),
						}
						var (
							found int
							got   []string
						)

						// NOTE: API doesn't guarantee DNS zone order.
						for k, v := range a {
							// Example of keys we're looking for:
							// "zones.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-1.fastly-example.com.",
							// "zones.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-2.fastly-example.com.",
							// "zones.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-3.fastly-example.com.",
							if strings.HasSuffix(k, ".name") {
								got = append(got, v)
								for _, name := range want {
									if v == name {
										found++
										break
									}
								}
							}
						}

						if found != len(want) {
							return fmt.Errorf("want: %v, got: %v", want, got)
						}

						return nil
					},
				),
			},
		},
	})
}

func testAccFastlyDataSourceDNSZonesConfig(h string) string {
	tf := `
resource "fastly_dns_zone" "example_1" {
  name = "tf-%s-1.fastly-example.com."
}

resource "fastly_dns_zone" "example_2" {
  name = "tf-%s-2.fastly-example.com."
}

resource "fastly_dns_zone" "example_3" {
  name = "tf-%s-3.fastly-example.com."
}

data "fastly_dns_zones" "example" {
  depends_on = [fastly_dns_zone.example_1, fastly_dns_zone.example_2, fastly_dns_zone.example_3]
}
`
	return fmt.Sprintf(tf, h, h, h)
}
