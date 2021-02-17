package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAccFastlyDataSourceTlSCertificateIDs(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSCertificateIDSConfig_resources(name, key, cert),
			},
			{
				Config: testAccFastlyDataSourceTLSCertificateIDSConfig_resourcesAndData(name, key, cert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("result", "true"),
				),
			},
		},
	})
}

func testAccFastlyDataSourceTLSCertificateIDSConfig_resources(name, key, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  key_pem = "%s"
  name = "%s"
}
resource "fastly_tls_certificate" "cert" {
  certificate_body = "%s"
  depends_on = [fastly_tls_private_key.key]
}
`, key, name, cert)
}

func testAccFastlyDataSourceTLSCertificateIDSConfig_resourcesAndData(name, key, cert string) string {
	return fmt.Sprintf(`
%s
data "fastly_tls_certificate_ids" "subject" {}
output "result" {
  value = tostring(contains(
    data.fastly_tls_certificate_ids.subject.ids, fastly_tls_certificate.cert.id
  ))
}
`, testAccFastlyDataSourceTLSCertificateIDSConfig_resources(name, key, cert))
}
