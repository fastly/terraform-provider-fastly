package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSPrivateKeyIds_basic(t *testing.T) {
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
				Config: testAccFastlyDataSourceTLSPrivateKeyIdsConfigOnlyTestKey(key, name),
			},
			{
				Config: testAccFastlyDataSourceTLSPrivateKeyIdsConfigTestKeyWithData(key, name),
				Check: resource.TestCheckTypeSetElemAttrPair(
					"data.fastly_tls_private_key_ids.subject", "ids.*",
					"fastly_tls_private_key.test", "id",
				),
			},
		},
	})
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
