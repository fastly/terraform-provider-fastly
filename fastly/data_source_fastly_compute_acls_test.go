package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyDataSourceACLs_Basic(t *testing.T) {
	h := generateHex()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceACLsConfig(h),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						r := s.RootModule().Resources["data.fastly_compute_acls.example"]
						a := r.Primary.Attributes

						want := generateNames(h, 3)
						var (
							found int
							got   []string
						)

						// NOTE: API doesn't guarantee ACLs order.
						for k, v := range a {
							// Example of keys we're looking for:
							// "acls.0.name":"tf_677f63804c9351ac31fd0cb1db697b95_1",
							// "acls.1.name":"tf_677f63804c9351ac31fd0cb1db697b95_2",
							// "acls.2.name":"tf_677f63804c9351ac31fd0cb1db697b95_3",
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

func testAccFastlyDataSourceACLsConfig(h string) string {
	tf := `
resource "fastly_compute_acl" "example_1" {
  name          = "tf_%s_1"
  force_destroy = true
}

resource "fastly_compute_acl" "example_2" {
  name          = "tf_%s_2"
  force_destroy = true
}

resource "fastly_compute_acl" "example_3" {
  name          = "tf_%s_3"
  force_destroy = true
}

data "fastly_compute_acls" "example" {
  depends_on = [fastly_compute_acl.example_1, fastly_compute_acl.example_2, fastly_compute_acl.example_3]
}
`

	return fmt.Sprintf(tf, h, h, h)
}
