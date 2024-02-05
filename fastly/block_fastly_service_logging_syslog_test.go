package fastly

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenSyslog(t *testing.T) {
	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("failed to generate key and cert: %s", err)
	}

	cases := []struct {
		remote []*gofastly.Syslog
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Syslog{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("somesyslogname"),
					Address:           gofastly.ToPointer("127.0.0.1"),
					IPV4:              gofastly.ToPointer("127.0.0.1"),
					Port:              gofastly.ToPointer(8080),
					Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
					FormatVersion:     gofastly.ToPointer(1),
					ResponseCondition: gofastly.ToPointer("response_condition_test"),
					MessageType:       gofastly.ToPointer("classic"),
					Token:             gofastly.ToPointer("abcd1234"),
					UseTLS:            gofastly.ToPointer(true),
					TLSCACert:         gofastly.ToPointer(cert),
					TLSHostname:       gofastly.ToPointer("example.com"),
					TLSClientCert:     gofastly.ToPointer(cert),
					TLSClientKey:      gofastly.ToPointer(key),
				},
			},
			local: []map[string]any{
				{
					"name":               "somesyslogname",
					"address":            "127.0.0.1",
					"port":               8080,
					"format":             `%h %l %u %t "%r" %>s %b`,
					"format_version":     1,
					"response_condition": "response_condition_test",
					"message_type":       "classic",
					"token":              "abcd1234",
					"use_tls":            true,
					"tls_hostname":       "example.com",
					"tls_ca_cert":        cert,
					"tls_client_cert":    cert,
					"tls_client_key":     key,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSyslogs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_syslog_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Syslog{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somesyslogname"),
		Address:           gofastly.ToPointer("127.0.0.1"),
		IPV4:              gofastly.ToPointer("127.0.0.1"),
		Port:              gofastly.ToPointer(514),
		Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		MessageType:       gofastly.ToPointer("classic"),
	}

	log1AfterUpdate := gofastly.Syslog{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somesyslogname"),
		Address:           gofastly.ToPointer("127.0.0.1"),
		IPV4:              gofastly.ToPointer("127.0.0.1"),
		Port:              gofastly.ToPointer(514),
		Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		MessageType:       gofastly.ToPointer("blank"),
	}

	log2 := gofastly.Syslog{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("somesyslogname2"),
		Address:        gofastly.ToPointer("127.0.0.2"),
		IPV4:           gofastly.ToPointer("127.0.0.2"),
		Port:           gofastly.ToPointer(10514),
		Format:         gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:  gofastly.ToPointer(2),
		MessageType:    gofastly.ToPointer("classic"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSyslogConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSyslogAttributes(&service, []*gofastly.Syslog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_syslog.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSyslogConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSyslogAttributes(&service, []*gofastly.Syslog{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_syslog.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_syslog_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Syslog{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("somesyslogname"),
		Address:        gofastly.ToPointer("127.0.0.1"),
		IPV4:           gofastly.ToPointer("127.0.0.1"),
		Port:           gofastly.ToPointer(514),
		MessageType:    gofastly.ToPointer("classic"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSyslogComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLSyslogAttributes(&service, []*gofastly.Syslog{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_syslog.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_syslog_formatVersion(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Syslog{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("somesyslogname"),
		Address:        gofastly.ToPointer("127.0.0.1"),
		IPV4:           gofastly.ToPointer("127.0.0.1"),
		Port:           gofastly.ToPointer(514),
		Format:         gofastly.ToPointer("%a %l %u %t %m %U%q %H %>s %b %T"),
		FormatVersion:  gofastly.ToPointer(2),
		MessageType:    gofastly.ToPointer("classic"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSyslogConfigFormatVersion(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSyslogAttributes(&service, []*gofastly.Syslog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_syslog.#", "1"),
				),
			},
		},
	})
}

// This test should not be run in parallel due to its use of schema.EnvDefaultFunc to set/reset environment variables,
// which conflicts with other running tests.
func TestAccFastlyServiceVCL_syslog_useTLS(t *testing.T) {
	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("failed to generate key and cert: %s", err)
	}
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// set env Vars to something we expect
	resetEnv := setSyslogEnv(key, cert, t)
	defer resetEnv()

	log1 := gofastly.Syslog{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("somesyslogname"),
		Address:        gofastly.ToPointer("127.0.0.1"),
		IPV4:           gofastly.ToPointer("127.0.0.1"),
		Port:           gofastly.ToPointer(514),
		Format:         gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:  gofastly.ToPointer(2),
		MessageType:    gofastly.ToPointer("classic"),
		UseTLS:         gofastly.ToPointer(true),
		TLSCACert:      gofastly.ToPointer(cert),
		TLSHostname:    gofastly.ToPointer("example.com"),
		TLSClientCert:  gofastly.ToPointer(cert),
		TLSClientKey:   gofastly.ToPointer(key),
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSyslogConfigUseTLS(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSyslogAttributes(&service, []*gofastly.Syslog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_syslog.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLSyslogAttributes(service *gofastly.ServiceDetail, syslogs []*gofastly.Syslog, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		syslogList, err := conn.ListSyslogs(&gofastly.ListSyslogsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Syslog Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(syslogList) != len(syslogs) {
			return fmt.Errorf("syslog List count mismatch, expected (%d), got (%d)", len(syslogs), len(syslogList))
		}

		log.Printf("[DEBUG] syslogList = %+v\n", syslogList)

		var found int
		for _, s := range syslogs {
			for _, ls := range syslogList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(ls.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					ls.CreatedAt = nil
					ls.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						ls.FormatVersion = s.FormatVersion
						ls.Format = s.Format
						ls.ResponseCondition = s.ResponseCondition
						ls.Placement = s.Placement
					}

					if !reflect.DeepEqual(s, ls) {
						return fmt.Errorf("bad match Syslog logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", s, ls)
					}
					found++
				}
			}
		}

		if found != len(syslogs) {
			return fmt.Errorf("error matching Syslog Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLSyslogComputeConfig(name, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  logging_syslog {
    name               = "somesyslogname"
    address            = "127.0.0.1"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSyslogConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }
  logging_syslog {
    name               = "somesyslogname"
    address            = "127.0.0.1"
    response_condition = "response_condition_test"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSyslogConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }
  logging_syslog {
    name               = "somesyslogname"
    address            = "127.0.0.1"
    port               = 514
    response_condition = "response_condition_test"
    message_type       = "blank"
  }
  logging_syslog {
    name    = "somesyslogname2"
    address = "127.0.0.2"
    port    = 10514
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSyslogConfigFormatVersion(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  logging_syslog {
    name           = "somesyslogname"
    address        = "127.0.0.1"
    port           = 514
    format         = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
    format_version = 2
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSyslogConfigUseTLS(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  logging_syslog {
    address            = "127.0.0.1"
    name               = "somesyslogname"
    port               = 514
    tls_hostname       = "example.com"
    use_tls            = true
  }
  force_destroy = true
}`, name, domain)
}

func setSyslogEnv(key string, cert string, t *testing.T) func() {
	e := getSyslogEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_SYSLOG_CA_CERT", cert); err != nil {
		t.Fatalf("Error setting env var FASTLY_SYSLOG_CA_CERT: %s", err)
	}
	if err := os.Setenv("FASTLY_SYSLOG_CLIENT_CERT", cert); err != nil {
		t.Fatalf("Error setting env var FASTLY_SYSLOG_CLIENT_CERT: %s", err)
	}
	if err := os.Setenv("FASTLY_SYSLOG_CLIENT_KEY", key); err != nil {
		t.Fatalf("Error setting env var FASTLY_SYSLOG_CLIENT_KEY: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_SYSLOG_CA_CERT", e.CaCert); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SYSLOG_CA_CERT: %s", err)
		}
		if err := os.Setenv("FASTLY_SYSLOG_CLIENT_CERT", e.ClientCert); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SYSLOG_CLIENT_CERT: %s", err)
		}
		if err := os.Setenv("FASTLY_SYSLOG_CLIENT_KEY", e.ClientKey); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SYSLOG_CLIENT_KEY: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentSyslogEnv struct {
	CaCert, ClientCert, ClientKey string
}

func getSyslogEnv() *currentSyslogEnv {
	// Grab any existing Fastly Syslog certs and keys and preserve, in the off chance
	// they're actually set in the environment
	return &currentSyslogEnv{
		CaCert:     os.Getenv("FASTLY_SYSLOG_CA_CERT"),
		ClientCert: os.Getenv("FASTLY_SYSLOG_CLIENT_CERT"),
		ClientKey:  os.Getenv("FASTLY_SYSLOG_CLIENT_KEY"),
	}
}
