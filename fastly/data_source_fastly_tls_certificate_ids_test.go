package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTlSCertificateIDs(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.com", acctest.RandomWithPrefix(testResourcePrefix))

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSCertificateIDSConfigResources(name, key, cert),
				Check: resource.TestCheckTypeSetElemAttrPair(
					"data.fastly_tls_certificate_ids.subject", "ids.*",
					"fastly_tls_certificate.cert", "id"),
			},
		},
	})
}

func testAccFastlyDataSourceTLSCertificateIDSConfigResources(name, key, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  key_pem = "%s"
  name = "%s"
}
resource "fastly_tls_certificate" "cert" {
  certificate_body = "%s"
  depends_on = [fastly_tls_private_key.key]
}
data "fastly_tls_certificate_ids" "subject" {
  depends_on = [fastly_tls_certificate.cert]
}
`, key, name, cert)
}
