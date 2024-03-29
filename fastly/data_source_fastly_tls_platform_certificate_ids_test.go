package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSPlatformCertificate_IDs(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.test", name)

	key, cert, ca, err := generateKeyAndCertWithCA(domain)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSPlatformCertificateIDSConfigResources(name, key, cert, ca),
			},
			{
				Config: testAccFastlyDataSourceTLSPlatformCertificateIDSConfigResourcesAndData(name, key, cert, ca),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(
						"data.fastly_tls_platform_certificate_ids.subject", "ids.*",
						"fastly_tls_platform_certificate.cert", "id",
					),
				),
			},
		},
	})
}

func testAccFastlyDataSourceTLSPlatformCertificateIDSConfigResources(name, key, cert, ca string) string {
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
`, name, key, cert, ca)
}

func testAccFastlyDataSourceTLSPlatformCertificateIDSConfigResourcesAndData(name, key, cert, ca string) string {
	return fmt.Sprintf(`
%s
data "fastly_tls_platform_certificate_ids" "subject" {}
`, testAccFastlyDataSourceTLSPlatformCertificateIDSConfigResources(name, key, cert, ca))
}
