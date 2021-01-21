package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenSplunk(t *testing.T) {
	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("Failed to generate key and cert: %s", err)
	}

	cases := []struct {
		remote []*gofastly.Splunk
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Splunk{
				{
					Name:              "test-splunk",
					URL:               "https://mysplunkendpoint.example.com/services/collector/event",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     1,
					ResponseCondition: "error_response",
					Placement:         "waf_debug",
					Token:             "test-token",
					// The same certificate is used here for
					// TLSCACert and TLSClientCert, but this
					// is strictly for testing. In practice
					// the same value should not be used for
					// these two fields.
					TLSCACert:     cert,
					TLSHostname:   "example.com",
					TLSClientCert: cert,
					TLSClientKey:  key,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "test-splunk",
					"url":                "https://mysplunkendpoint.example.com/services/collector/event",
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(1),
					"response_condition": "error_response",
					"placement":          "waf_debug",
					"token":              "test-token",
					"tls_hostname":       "example.com",
					// The same certificate is used here for
					// TLSCACert and TLSClientCert, but this
					// is strictly for testing. In practice
					// the same value should not be used for
					// these two fields.
					"tls_ca_cert":     cert,
					"tls_client_cert": cert,
					"tls_client_key":  key,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSplunks(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_splunk_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLogOne := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	splunkLogOneUpdated := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	splunkLogTwo := gofastly.Splunk{
		Name:              "test-splunk-2",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "ok_response_2XX",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_basic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},

			{
				Config: testAccServiceV1SplunkConfig_update(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOneUpdated, &splunkLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLogOne := gofastly.Splunk{
		Name:  "test-splunk-1",
		URL:   "https://mysplunkendpoint.example.com/services/collector/event",
		Token: "test-token",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfigCompute_basic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "splunk.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLog := gofastly.Splunk{
		Name:          "test-splunk",
		URL:           "https://mysplunkendpoint.example.com/services/collector/event",
		Token:         "test-token",
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_default(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLog}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_complete(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("Failed to generate key and cert: %s", err)
	}

	splunkLogOne := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
		TLSHostname:       "example.com",
		// The same certificate is used here for
		// TLSCACert and TLSClientCert, but this
		// is strictly for testing. In practice
		// the same value should not be used for
		// these two fields.
		TLSCACert:     cert,
		TLSClientCert: cert,
		TLSClientKey:  key,
	}

	splunkLogOneUpdated := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
		TLSHostname:       "example.com",
		// The same certificate is used here for
		// TLSCACert and TLSClientCert, but this
		// is stricly for testing. In practice
		// the same value should not be used for
		// these two fields.
		TLSCACert:     cert,
		TLSClientCert: cert,
		TLSClientKey:  key,
	}

	splunkLogTwo := gofastly.Splunk{
		Name:              "test-splunk-2",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "ok_response_2XX",
		TLSHostname:       "example.com",
		// The same certificate is used here for
		// TLSCACert and TLSClientCert, but this
		// is stricly for testing. In practice
		// the same value should not be used for
		// these two fields.
		TLSCACert:     cert,
		TLSClientCert: cert,
		TLSClientKey:  key,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_useTLS(serviceName, cert, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},

			{
				Config: testAccServiceV1SplunkConfig_updateUseTLS(serviceName, cert, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOneUpdated, &splunkLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_env(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	_, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("Failed to generate key and cert: %s", err)
	}

	// set env variable to something we expect
	resetEnv := setSplunkEnv("test-token", cert, t)
	defer resetEnv()

	splunkLog := gofastly.Splunk{
		Name:          "test-splunk",
		URL:           "https://mysplunkendpoint.example.com/services/collector/event",
		Token:         "test-token",
		TLSCACert:     cert,
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_env(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLog}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1SplunkAttributes(service *gofastly.ServiceDetail, localSplunkList []*gofastly.Splunk, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		remoteSplunkList, err := conn.ListSplunks(&gofastly.ListSplunksInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Splunk for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(remoteSplunkList) != len(localSplunkList) {
			return fmt.Errorf("Splunk List count mismatch, expected (%d), got (%d)", len(localSplunkList), len(remoteSplunkList))
		}

		var found int
		for _, ls := range localSplunkList {
			for _, rs := range remoteSplunkList {
				if ls.Name == rs.Name {
					// we don't know these things ahead of time, so populate them now
					ls.ServiceID = service.ID
					ls.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					rs.CreatedAt = nil
					rs.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						rs.FormatVersion = ls.FormatVersion
						rs.Format = ls.Format
						rs.ResponseCondition = ls.ResponseCondition
						rs.Placement = ls.Placement
					}

					if !reflect.DeepEqual(ls, rs) {
						return fmt.Errorf("Bad match Splunk logging match, expected (%#v), got (%#v)", ls, rs)
					}
					found++
				}
			}
		}

		if found != len(localSplunkList) {
			return fmt.Errorf("Error matching Splunk rules")
		}

		return nil
	}
}

func testAccServiceV1SplunkConfigCompute_basic(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}


  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1SplunkConfig_basic(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	format := "%h %l %u %t \"%r\" %>s %b"

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 1
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  force_destroy = true
}`, serviceName, domainName, format)
}

func testAccServiceV1SplunkConfig_useTLS(serviceName, cert, key string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	format := "%h %l %u %t \"%r\" %>s %b"

	// The same certificate is used here for tls_ca_cert and tls_client_cert,
	// but this is stricly for testing. In practice the same value should
	// not be used for these two fields.
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 1
    placement          = "waf_debug"
    tls_hostname       = "example.com"
    tls_ca_cert        = %q
    tls_client_cert    = %q
    tls_client_key     = %q
    response_condition = "error_response_5XX"
  }

  force_destroy = true
}`, serviceName, domainName, format, cert, cert, key)
}

func testAccServiceV1SplunkConfig_update(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	format := "%h %l %u %%{now}V %%{req.method}V %%{req.url}V %>s %%{resp.http.Content-Length}V"

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  condition {
    name      = "ok_response_2XX"
    statement = "resp.status >= 200 && resp.status < 300"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 2
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  splunk {
    name               = "test-splunk-2"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 2
    placement          = "waf_debug"
    response_condition = "ok_response_2XX"
  }

  force_destroy = true
}`, serviceName, domainName, format, format)
}

func testAccServiceV1SplunkConfig_updateUseTLS(serviceName, cert, key string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	format := "%h %l %u %%{now}V %%{req.method}V %%{req.url}V %>s %%{resp.http.Content-Length}V"

	// The same certificate is used here for tls_ca_cert and tls_client_cert,
	// but this is stricly for testing. In practice the same value should
	// not be used for these two fields.
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  condition {
    name      = "ok_response_2XX"
    statement = "resp.status >= 200 && resp.status < 300"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 2
    placement          = "waf_debug"
    tls_hostname       = "example.com"
    tls_ca_cert        = %q
    tls_client_cert    = %q
    tls_client_key     = %q
    response_condition = "error_response_5XX"
  }

  splunk {
    name               = "test-splunk-2"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 2
    placement          = "waf_debug"
    tls_hostname       = "example.com"
    tls_ca_cert        = %q
    tls_client_cert    = %q
    tls_client_key     = %q
    response_condition = "ok_response_2XX"
  }

  force_destroy = true
}`, serviceName, domainName, format, cert, cert, key, format, cert, cert, key)
}

func testAccServiceV1SplunkConfig_default(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  splunk {
    name  = "test-splunk"
    url   = "https://mysplunkendpoint.example.com/services/collector/event"
    token = "test-token"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1SplunkConfig_env(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  splunk {
    name = "test-splunk"
    url  = "https://mysplunkendpoint.example.com/services/collector/event"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func setSplunkEnv(token string, cert string, t *testing.T) func() {
	e := getSplunkEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_SPLUNK_TOKEN", token); err != nil {
		t.Fatalf("Error setting env var FASTLY_SPLUNK_TOKEN: %s", err)
	}

	if err := os.Setenv("FASTLY_SPLUNK_CA_CERT", cert); err != nil {
		t.Fatalf("Error setting env var FASTLY_SPLUNK_CA_CERT: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_SPLUNK_TOKEN", e.Token); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SPLUNK_TOKEN: %s", err)
		}

		if err := os.Setenv("FASTLY_SPLUNK_CA_CERT", e.Token); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SPLUNK_CA_CERT: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentSplunkEnv struct {
	Token, CaCert string
}

func getSplunkEnv() *currentSplunkEnv {
	// Grab the existing Fastly Splunk-related environment variables and preserve,
	// in the off chance they're actually set in the environment.
	return &currentSplunkEnv{
		Token:  os.Getenv("FASTLY_SPLUNK_TOKEN"),
		CaCert: os.Getenv("FASTLY_SPLUNK_CA_CERT"),
	}
}
