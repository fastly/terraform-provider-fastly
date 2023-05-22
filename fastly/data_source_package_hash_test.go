package fastly

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyPackageHash_Config(t *testing.T) {
	validPackageContent, _ := os.ReadFile("./test_fixtures/package/valid.tar.gz")
	b64Content := base64.StdEncoding.EncodeToString(validPackageContent)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: `
        data "fastly_package_hash" "example" {
          filename = "./test_fixtures/package/valid.tar.gz"
        }
        `,
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyPackageHashState("data.fastly_package_hash.example"),
				),
			},
			{
				Config: fmt.Sprintf(`
        data "fastly_package_hash" "example" {
          content = "%s"
        }
        `, b64Content),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyPackageHashState("data.fastly_package_hash.example"),
				),
			},
		},
	})
}

func testAccFastlyPackageHashState(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[n]
		a := r.Primary.Attributes
		if a["hash"] != "a763d3c88968ebc17691900d3c14306762296df8e47a1c2d7661cee0e0c5aa6d4c082a7c128d6e719fe333b73b46fe3ae32694716ccd2efa21f5d9f049ceec6d" {
			return fmt.Errorf("unexpected package hash: %s", a["hash"])
		}
		return nil
	}
}
