package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSPrivateKeyFilters(t *testing.T) {
	key, _, err := generateKeyAndCert()
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)

	name := acctest.RandomWithPrefix(testResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyConfigByName(key, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "name", name),
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "key_type", "RSA"),
				),
			},
		},
	})
}

func TestAccFastlyDataSourceTLSPrivateKeyByID(t *testing.T) {
	key, _, err := generateKeyAndCert()
	require.NoError(t, err)

	name := acctest.RandomWithPrefix(testResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyConfigByID(key, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "name", name),
					resource.TestCheckResourceAttr("data.fastly_tls_private_key.subject", "key_type", "RSA"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceTLSPrivateKeyConfigByName(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = "%s"
  name = "%s"
}

data "fastly_tls_private_key" "subject" {
  name = fastly_tls_private_key.test.name
}
`, key, name)
}

func testAccFastlyDataSourceTLSPrivateKeyConfigByID(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = <<EOF
%s
EOF
  name = "%s"
}

data "fastly_tls_private_key" "subject" {
  id = fastly_tls_private_key.test.id
}
`, key, name)
}
