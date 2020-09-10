package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccFastlyACLEntries(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyACLEntriesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyACLEntries("data.fastly_acl_entries.some"),
				),
			},
		},
	})
}

func testAccFastlyACLEntries(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes

		attrsToTest := map[string]interface{}{
			"id":      "abc",
			"ip":      "127.0.0.1",
			"subnet":  "32",
			"negated": false,
			"comment": "",
		}

		for attr, expectedValue := range attrsToTest {
			if a[attr] != expectedValue {
				return fmt.Errorf(
					"%s is %s; want %s",
					attr,
					a[attr],
					expectedValue,
				)
			}
		}
		return nil
	}
}

const testAccFastlyACLEntriesConfig = `
data "fastly_acl_entries" "some" {
	service = "123456789"
    acl     = "abcde12345"
}
`
