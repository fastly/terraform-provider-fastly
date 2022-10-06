package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenSplunk(t *testing.T) {
	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("failed to generate key and cert: %s", err)
	}

	cases := []struct {
		remote []*gofastly.Splunk
		local  []map[string]any
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
			local: []map[string]any{
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
					"use_tls":         false,
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

func TestAccFastlyServiceVCL_splunk_basic(t *testing.T) {
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
		UseTLS:            true,
	}

	splunkLogOneUpdated := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
		UseTLS:            false,
	}

	splunkLogTwo := gofastly.Splunk{
		Name:              "test-splunk-2",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "ok_response_2XX",
		UseTLS:            false,
	}

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSplunkConfigBasic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_splunk.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSplunkConfigUpdate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOneUpdated, &splunkLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_splunk.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_splunk_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLogOne := gofastly.Splunk{
		Name:  "test-splunk-1",
		URL:   "https://mysplunkendpoint.example.com/services/collector/event",
		Token: "test-token",
	}

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSplunkConfigComputeBasic(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_splunk.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_splunk_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLog := gofastly.Splunk{
		Name:          "test-splunk",
		URL:           "https://mysplunkendpoint.example.com/services/collector/event",
		Token:         "test-token",
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 2,
	}

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSplunkConfigDefault(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLog}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_splunk.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_splunk_complete(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	key, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("failed to generate key and cert: %s", err)
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
		// is strictly for testing. In practice
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
		// is strictly for testing. In practice
		// the same value should not be used for
		// these two fields.
		TLSCACert:     cert,
		TLSClientCert: cert,
		TLSClientKey:  key,
	}

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSplunkConfigUseTLS(serviceName, cert, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_splunk.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSplunkConfigUpdateUseTLS(serviceName, cert, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOneUpdated, &splunkLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_splunk.#", "2"),
				),
			},
		},
	})
}

func TestSplunkEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingSplunk(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_splunk"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["token"].Sensitive {
		t.Fatalf("Expected token to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	_, cert, err := generateKeyAndCert()
	if err != nil {
		t.Errorf("failed to generate key and cert: %s", err)
	}
	token := "test-token"
	resetEnv := setSplunkEnv(cert, token, t)
	defer resetEnv()

	result1, err1 := loggingResourceSchema["token"].DefaultFunc()
	if err1 != nil {
		t.Fatalf("Unexpected err %#v when calling token DefaultFunc", err1)
	}
	if result1 != token {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", token, result1)
	}

	result2, err2 := loggingResourceSchema["tls_ca_cert"].DefaultFunc()
	if err2 != nil {
		t.Fatalf("Unexpected err %#v when calling tls_ca_cert DefaultFunc", err2)
	}
	if result2 != cert {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", cert, result2)
	}
}

func testAccCheckFastlyServiceVCLSplunkAttributes(service *gofastly.ServiceDetail, localSplunkList []*gofastly.Splunk, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		remoteSplunkList, err := conn.ListSplunks(&gofastly.ListSplunksInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Splunk for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(remoteSplunkList) != len(localSplunkList) {
			return fmt.Errorf("splunk List count mismatch, expected (%d), got (%d)", len(localSplunkList), len(remoteSplunkList))
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
						return fmt.Errorf("bad match Splunk logging match, expected (%#v), got (%#v)", ls, rs)
					}
					found++
				}
			}
		}

		if found != len(localSplunkList) {
			return fmt.Errorf("error matching Splunk rules")
		}

		return nil
	}
}

func testAccServiceVCLSplunkConfigComputeBasic(serviceName string) string {
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

  logging_splunk {
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

func testAccServiceVCLSplunkConfigBasic(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	format := "%h %l %u %t \"%r\" %>s %b"

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  logging_splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 1
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
	use_tls            = true
  }

  force_destroy = true
}`, serviceName, domainName, format)
}

func testAccServiceVCLSplunkConfigUseTLS(serviceName, cert, key string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	format := "%h %l %u %t \"%r\" %>s %b"

	// The same certificate is used here for tls_ca_cert and tls_client_cert,
	// but this is strictly for testing. In practice the same value should
	// not be used for these two fields.
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  logging_splunk {
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

func testAccServiceVCLSplunkConfigUpdate(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	format := "%h %l %u %%{now}V %%{req.method}V %%{req.url}V %>s %%{resp.http.Content-Length}V"

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  logging_splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = %q
    format_version     = 2
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  logging_splunk {
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

func testAccServiceVCLSplunkConfigUpdateUseTLS(serviceName, cert, key string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	format := "%h %l %u %%{now}V %%{req.method}V %%{req.url}V %>s %%{resp.http.Content-Length}V"

	// The same certificate is used here for tls_ca_cert and tls_client_cert,
	// but this is strictly for testing. In practice the same value should
	// not be used for these two fields.
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  logging_splunk {
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

  logging_splunk {
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

func testAccServiceVCLSplunkConfigDefault(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  logging_splunk {
    name  = "test-splunk"
    url   = "https://mysplunkendpoint.example.com/services/collector/event"
    token = "test-token"
    }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceVCLSplunkConfigEnv(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = %q

  domain {
    name    = %q
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  logging_splunk {
    name = "test-splunk"
    url  = "https://mysplunkendpoint.example.com/services/collector/event"
  }

  force_destroy = true
}`, serviceName, domainName)
}

// setSplunkEnv sets the specified values as environment variables and returns a
// function that can be used to reset the environment variables in case the
// same variables happen to be in the user's environment when running the tests.
func setSplunkEnv(cert, token string, t *testing.T) func() {
	e := getSplunkEnv()

	// stub specified environment variables
	if err := os.Setenv("FASTLY_SPLUNK_CA_CERT", cert); err != nil {
		t.Fatalf("Error setting env var FASTLY_SPLUNK_CA_CERT: %s", err)
	}
	if err := os.Setenv("FASTLY_SPLUNK_TOKEN", token); err != nil {
		t.Fatalf("Error setting env var FASTLY_SPLUNK_TOKEN: %s", err)
	}

	// function will reset all the environment variables we modified above
	return func() {
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
