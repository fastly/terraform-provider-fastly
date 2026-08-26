package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceTSIGKeys_Config(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTSIGKeysConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_tsig_keys.example"]
						a := r.Primary.Attributes

						want := []string{
							fmt.Sprintf("tf-%s-1.fastly-example.com", h),
							fmt.Sprintf("tf-%s-2.fastly-example.com", h),
							fmt.Sprintf("tf-%s-3.fastly-example.com", h),
						}
						var (
							found int
							got   []string
						)

						// NOTE: API doesn't guarantee TSIG key order.
						for k, v := range a {
							// Example of keys we're looking for:
							// "keys.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-1.fastly-example.com",
							// "keys.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-2.fastly-example.com",
							// "keys.1234567890.name":"tf-677f63804c9351ac31fd0cb1db697b95-3.fastly-example.com",
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

func testAccFastlyDataSourceTSIGKeysConfig(h string) string {
	tf := `
resource "fastly_tsig_key" "example_1" {
  name      = "tf-%s-1.fastly-example.com"
  algorithm = "hmac-sha256"
  secret    = "dGVzdHNlY3JldA=="
}

resource "fastly_tsig_key" "example_2" {
  name      = "tf-%s-2.fastly-example.com"
  algorithm = "hmac-sha256"
  secret    = "dGVzdHNlY3JldA=="
}

resource "fastly_tsig_key" "example_3" {
  name      = "tf-%s-3.fastly-example.com"
  algorithm = "hmac-sha256"
  secret    = "dGVzdHNlY3JldA=="
}

data "fastly_tsig_keys" "example" {
  depends_on = [fastly_tsig_key.example_1, fastly_tsig_key.example_2, fastly_tsig_key.example_3]
}
`
	return fmt.Sprintf(tf, h, h, h)
}
