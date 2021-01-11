package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"errors"
	"os"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyTLSActivationV1_basic(t *testing.T) {
	var ta gofastly.TLSActivation
	// I am looking for guidance here. Because there is no create API for TLSConfiguration we cannot create with terraform a new TLSConfiguration.
	// These tests work when I replace the next line with a TLSConfiguration ID from my account but of course that won't work for CI/CD purposes.
	// We can call the API from here to GET a list of configurations and just use the 1st one OR I can add a terraform data source for TLSConfiguration.
	// This would also require adding TLSConfiguration List and Get operations to the go-fastly package.
	configurationID, err := getFastlyTestingTLSConfigurationID()
	if err != nil {
		t.Fatal(err)
	}

	hostname := fmt.Sprintf("%s.com", acctest.RandString(10))
	key, certBlob, err := generateKeyAndCertWithSan(hostname)
	if err != nil {
		t.Fatal(err)
	}

	certBlob = strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTLSActivationV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSActivationV1(key, name, certBlob, hostname, configurationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSActivationV1Exists("fastly_tls_activation_v1.foo", &ta),
					resource.TestCheckResourceAttrSet(
						"fastly_tls_activation_v1.foo", "certificate_id"),
					resource.TestCheckResourceAttrSet(
						"fastly_tls_activation_v1.foo", "configuration_id"),
					resource.TestCheckResourceAttr(
						"fastly_tls_activation_v1.foo", "domain_id", hostname),
				),
			},
		},
	})
}

func TestAccFastlyTLSActivationV1_updateCertificateID(t *testing.T) {
	var ta gofastly.TLSActivation
	configurationID, err := getFastlyTestingTLSConfigurationID()
	if err != nil {
		t.Fatal(err)
	}
	hostname := fmt.Sprintf("%s.com", acctest.RandString(10))
	key, certBlob, err := generateKeyAndCertWithSan(hostname)
	if err != nil {
		t.Fatal(err)
	}
	certBlob = strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	key2, certBlob2, err := generateKeyAndCertWithSan(hostname)
	if err != nil {
		t.Fatal(err)
	}
	certBlob2 = strings.ReplaceAll(certBlob2, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key2 = strings.ReplaceAll(key2, "\n", `\n`)

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTLSActivationV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSActivationV1(key, name, certBlob, hostname, configurationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSActivationV1Exists("fastly_tls_activation_v1.foo", &ta),
				),
			},
			{
				Config: testAccTLSActivationV1_updateCertificateID(key, name, certBlob, hostname, configurationID, key2, certBlob2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSActivationV1Exists("fastly_tls_activation_v1.foo", &ta),
				),
			},
		},
	})
}

func TestAccFastlyTLSActivationV1_import(t *testing.T) {
	var ta gofastly.TLSActivation
	configurationID, err := getFastlyTestingTLSConfigurationID()
	if err != nil {
		t.Fatal(err)
	}
	hostname := fmt.Sprintf("%s.com", acctest.RandString(10))
	key, certBlob, err := generateKeyAndCertWithSan(hostname)
	if err != nil {
		t.Fatal(err)
	}

	certBlob = strings.ReplaceAll(certBlob, "\n", `\n`) // This is needed in order to make the certificate a single line string with \n linebreaks
	key = strings.ReplaceAll(key, "\n", `\n`)
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFastlyCustomTLSCertificateV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTLSActivationV1(key, name, certBlob, hostname, configurationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTLSActivationV1Exists("fastly_tls_activation_v1.foo", &ta),
				),
			},
			{
				ResourceName:      "fastly_tls_activation_v1.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTLSActivationV1Exists(n string, ta *gofastly.TLSActivation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No TLS Activation ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latest, err := conn.GetTLSActivation(&gofastly.GetTLSActivationInput{
			ID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*ta = *latest

		return nil
	}
}

func testAccCheckTLSActivationV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_tls_activation_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListTLSActivations(&gofastly.ListTLSActivationsInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing TLS Activations when deleting Fastly TLS Activation (%s): %s", rs.Primary.ID, err)
		}

		for _, ta := range l {
			if ta.ID == rs.Primary.ID {
				// user still found
				return fmt.Errorf("[WARN] Tried deleting TLS Activation (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccTLSActivationV1(key, name, certBlob, hostname, configurationID string) string {
	return fmt.Sprintf(`
resource "fastly_private_key_v1" "foo" {
		key = "%s"
		name  = "%s"
	}
	resource "fastly_custom_tls_certificate_v1" "foo" {
		cert_blob = "%s"
		name  = fastly_private_key_v1.foo.name
	}
	resource "fastly_service_v1" "foo" {
  		name = "%s"
		domain {
			name    = "%s"
			comment = "tf-testing-domain"
		}
		backend {
			address = "example.com"
			name    = "tf-testing-backend"
		 }
  		force_destroy = true
	}
	resource "fastly_tls_activation_v1" "foo" {
		certificate_id = fastly_custom_tls_certificate_v1.foo.id
		configuration_id  = "%s"
		domain_id  = fastly_service_v1.foo.domain.*.name[0]
}`, key, name, certBlob, name, hostname, configurationID)
}

func testAccTLSActivationV1_updateCertificateID(key, name, certBlob, hostname, configurationID, key2, certBlob2 string) string {
	return fmt.Sprintf(`
resource "fastly_private_key_v1" "foo" {
		key = "%s"
		name  = "%s"
	}
	resource "fastly_custom_tls_certificate_v1" "foo" {
		cert_blob = "%s"
		name  = fastly_private_key_v1.foo.name
	}
	resource "fastly_service_v1" "foo" {
  		name = "%s"
		domain {
			name    = "%s"
			comment = "tf-testing-domain"
		}
		backend {
			address = "example.com"
			name    = "tf-testing-backend"
		 }
  		force_destroy = true
	}
	resource "fastly_tls_activation_v1" "foo" {
		certificate_id = fastly_custom_tls_certificate_v1.bar.id
		configuration_id  = "%s"
		domain_id  = fastly_service_v1.foo.domain.*.name[0]
	}
	resource "fastly_private_key_v1" "bar" {
		key = "%s"
		name  = "%s"
	}
	resource "fastly_custom_tls_certificate_v1" "bar" {
		cert_blob = "%s"
		name  = fastly_private_key_v1.bar.name
}`, key, name, certBlob, name, hostname, configurationID, key2, name+"-2", certBlob2)
}

// Testing Fastly TLS Activation functions requires a valid TLS Activation ID to be set as an environment variable
func getFastlyTestingTLSConfigurationID() (string, error) {
	configurationID := os.Getenv("FASTLY_TESTING_TLS_CONFIGURATION_ID")
	if configurationID == "" {
		return "", errors.New("FASTLY_TESTING_TLS_CONFIGURATION_ID environment variable must be set in order to run TLS Activation tests")
	}
	return configurationID, nil
}
