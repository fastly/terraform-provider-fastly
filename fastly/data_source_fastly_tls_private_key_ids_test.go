package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSPrivateKeyIds_basic(t *testing.T) {
	key, _, err := generateKeyAndCert()
	require.NoError(t, err)

	name := acctest.RandomWithPrefix(testResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyIdsConfigOnlyTestKey(key, name),
			},
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyIdsConfigTestKeyWithData(key, name),
				Check:  testAccTLSPrivateKeyIDIncluded("data.fastly_tls_private_key_ids.subject", "fastly_tls_private_key.test"),
			},
		},
	})
}

// This can be replaced with `TestCheckTypeSetElemNestedAttrs` when using SDK 2.x
func testAccTLSPrivateKeyIDIncluded(dataSourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		d, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", dataSourceName)
		}

		for k, v := range d.Primary.Attributes {
			if k == "ids.#" {
				continue
			}
			if !strings.HasPrefix(k, "ids.") {
				continue
			}
			if v == r.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("unable to find private key %s in list of private key ids", r.Primary.ID)
	}
}

func testAccFastlyDataSourceTLSPrivateKeyIdsConfigOnlyTestKey(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = <<EOF
%s
EOF
  name = "%s"
}
`, key, name)
}

func testAccFastlyDataSourceTLSPrivateKeyIdsConfigTestKeyWithData(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = <<EOF
%s
EOF
  name = "%s"
}

data "fastly_tls_private_key_ids" "subject" {}
`, key, name)
}
