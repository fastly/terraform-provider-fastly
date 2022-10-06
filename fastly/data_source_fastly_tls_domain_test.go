package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSDomain_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example", name)

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories:         testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSDomainResources(key, cert, name),
			},
			{
				Config: testAccFastlyDataSourceTLSDomainWithDataSource(key, cert, domain, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.fastly_tls_domain.subject", "tls_certificate_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(
						"data.fastly_tls_domain.subject", "tls_certificate_ids.*",
						"fastly_tls_certificate.example", "id",
					),
				),
			},
		},
	})
}

func testAccFastlyDataSourceTLSDomainResources(key string, cert string, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  key_pem = <<EOF
%[1]s
EOF
  name = "%[3]s"
}
resource "fastly_tls_certificate" "example" {
  name = fastly_tls_private_key.key.name
  certificate_body = <<EOF
%[2]s
EOF
}
`, key, cert, name)
}

func testAccFastlyDataSourceTLSDomainWithDataSource(key string, cert string, domain string, name string) string {
	return fmt.Sprintf(`
%s
data "fastly_tls_domain" "subject" {
  domain = "%s"
}
`, testAccFastlyDataSourceTLSDomainResources(key, cert, name), domain)
}
