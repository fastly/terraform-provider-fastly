package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func TestAccFastlyCustomTLSCertificate_basic(t *testing.T) {
	var cc gofastly.CustomTLSCertificate
	san := fmt.Sprintf("%s.com", acctest.RandString(10))
	key, certBlob, err := generateKeyAndCertWithSan(san)
	if err != nil {
		t.Fatal(err)
	}
	singleLineCertBlob := strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	name := fmt.Sprintf("tf-test-cert-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastlyCustomTLSCertificateV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateKeyV1AndCustomTLSCertificateV1Config(key, singleLineCertBlob, name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyCustomTLSCertificateV1Exists("fastly_custom_tls_certificate_v1.foo", &cc),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "cert_blob", certBlob),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "name", fmt.Sprintf("%s-%s", name, name)),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "not_after"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "not_before"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "replace"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "serial_number"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "signature_algorithm"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "created_at"),
					resource.TestCheckResourceAttrSet(
						"fastly_custom_tls_certificate_v1.foo", "updated_at"),
					testAccCheckFastlyCustomTLSV1DomainsMatch("fastly_custom_tls_certificate_v1.foo", san),
				),
			},
		},
	})
}

func TestAccFastlyCustomTLSCertificate_updateName(t *testing.T) {
	var cc gofastly.CustomTLSCertificate
	key, certBlob, err := generateKeyAndCertWithSan(fmt.Sprintf("%s.com", acctest.RandString(10)))
	if err != nil {
		t.Fatal(err)
	}

	singleLineCertBlob := strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	keyName := fmt.Sprintf("tf-test-cert-%s", acctest.RandString(10))
	certName := fmt.Sprintf("tf-test-cert-%s", acctest.RandString(10))
	certName2 := fmt.Sprintf("tf-test-cert-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastlyCustomTLSCertificateV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateKeyV1AndCustomTLSCertificateV1Config(key, singleLineCertBlob, keyName, certName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyCustomTLSCertificateV1Exists("fastly_custom_tls_certificate_v1.foo", &cc),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "cert_blob", certBlob),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "name", fmt.Sprintf("%s-%s", keyName, certName))),
			},
			{
				Config: testAccPrivateKeyV1AndCustomTLSCertificateV1Config(key, singleLineCertBlob, keyName, certName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyCustomTLSCertificateV1Exists("fastly_custom_tls_certificate_v1.foo", &cc),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "name", fmt.Sprintf("%s-%s", keyName, certName2)),
				),
			},
		},
	})
}

func TestAccFastlyCustomTLSCertificateV1_import(t *testing.T) {
	var cc gofastly.CustomTLSCertificate
	key, certBlob, err := generateKeyAndCertWithSan(fmt.Sprintf("%s.com", acctest.RandString(10)))
	if err != nil {
		t.Fatal(err)
	}
	singleLineCertBlob := strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	name := fmt.Sprintf("tf-test-cert-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastlyCustomTLSCertificateV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateKeyV1AndCustomTLSCertificateV1Config(key, singleLineCertBlob, name, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFastlyCustomTLSCertificateV1Exists("fastly_custom_tls_certificate_v1.foo", &cc),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "cert_blob", certBlob),
					resource.TestCheckResourceAttr(
						"fastly_custom_tls_certificate_v1.foo", "name", fmt.Sprintf("%s-%s", name, name)),
				),
			},
			{
				ResourceName:            "fastly_custom_tls_certificate_v1.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cert_blob"},
			},
		},
	})
}

func testAccCheckFastlyCustomTLSCertificateV1Exists(n string, cc *gofastly.CustomTLSCertificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Custom TLS Certifcate ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latest, err := conn.GetCustomTLSCertificate(&gofastly.GetCustomTLSCertificateInput{
			ID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*cc = *latest

		return nil
	}
}

func testAccCheckFastlyCustomTLSCertificateV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_custom_tls_certificate_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListCustomTLSCertificates(&gofastly.ListCustomTLSCertificatesInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing custom TLS certificates when deleting Fastly custom TLS certificate (%s): %s", rs.Primary.ID, err)
		}

		for _, c := range l {
			if c.ID == rs.Primary.ID {
				// user still found
				return fmt.Errorf("[WARN] Tried deleting Custom TLS Certificate (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckFastlyCustomTLSV1DomainsMatch(n, expectedDomain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Custom TLS Certifcate ID is set")
		}

		domain, ok := rs.Primary.Attributes["domains.0"]
		if !ok {
			return fmt.Errorf("no domains found in state of Custom TLS Certificate")
		}
		if domain != expectedDomain {
			return fmt.Errorf("no domain matching %s found in state", domain)
		}

		return nil
	}
}

func testAccPrivateKeyV1AndCustomTLSCertificateV1Config(key, certBlob, keyName, certName string) string {
	return fmt.Sprintf(`
resource "fastly_private_key_v1" "foo" {
	key = "%s"
	name  = "%s"
}
resource "fastly_custom_tls_certificate_v1" "foo" {
	cert_blob = "%s"
	name  = "${fastly_private_key_v1.foo.name}-%s"
}`, key, keyName, certBlob, certName)
}
