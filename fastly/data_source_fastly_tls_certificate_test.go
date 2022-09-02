package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyDataSourceTLSCertificate_withName(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	dataSourceName := "data.fastly_tls_certificate.test"
	resourceName := "fastly_tls_certificate.cert"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTLSCertificate(name, key, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "created_at", resourceName, "created_at"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "updated_at", resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "issued_to", resourceName, "issued_to"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "issuer", resourceName, "issuer"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "replace", resourceName, "replace"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "serial_number", resourceName, "serial_number"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "signature_algorithm", resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "domains", resourceName, "domains"),
				),
			},
		},
	})
}

func testAccDataSourceTLSCertificate(keyName string, key string, cert string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

resource "fastly_tls_certificate" "cert" {
  name = "%[1]s"
  certificate_body = <<EOF
%[3]s
EOF
  depends_on = [fastly_tls_private_key.key]
}

data "fastly_tls_certificate" "test" {
  name = fastly_tls_certificate.cert.name
  domains = ["%[4]s"]
}
`, keyName, key, cert, domain)
}
