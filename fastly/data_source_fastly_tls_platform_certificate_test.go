package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSPlatformCertificate(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.test", name)

	key, cert, ca, err := generateKeyAndCertWithCA(domain)
	require.NoError(t, err)

	dataSourceName := "data.fastly_tls_platform_certificate.test"
	resourceName := "fastly_tls_platform_certificate.cert"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSPlatformCertificateConfig(name, key, cert, ca),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "created_at", resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "replace", resourceName, "replace"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "domains", resourceName, "domains"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "configuration_id", "data.fastly_tls_configuration.config", "id"),
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSPlatformCertificateConfig(keyName, key, cert, ca string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

data "fastly_tls_configuration" "config" {
  tls_service = "PLATFORM"
}

resource "fastly_tls_platform_certificate" "cert" {
  certificate_body = <<EOF
%[3]s
EOF
  intermediates_blob = <<EOF
%[4]s
EOF
  configuration_id = data.fastly_tls_configuration.config.id
  allow_untrusted_root = true
  depends_on = [fastly_tls_private_key.key]
}
data "fastly_tls_platform_certificate" "test" {
  id = fastly_tls_platform_certificate.cert.id
}
`, keyName, key, cert, ca)
}
