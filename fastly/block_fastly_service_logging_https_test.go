package fastly

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func TestAccFastlyServiceLoggingHTTPS_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.HTTPS{
		CompressionCodec:  new("zstd"),
		ContentType:       new(""),
		Format:            new(LoggingHTTPSDefaultFormat),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("PUT"),
		Name:              new("httpslogger"),
		Period:            new(10),
		RequestMaxBytes:   new(0),
		RequestMaxEntries: new(0),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/1"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.HTTPS{
		CompressionCodec:  new("snappy"),
		ContentType:       new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("POST"),
		Name:              new("httpslogger"),
		Period:            new(30),
		RequestMaxBytes:   new(0),
		RequestMaxEntries: new(0),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/1"),
		ProcessingRegion:  new("none"),
	}

	log1WithCondition := gofastly.HTTPS{
		CompressionCodec:  new("snappy"),
		ContentType:       new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("POST"),
		Name:              new("httpslogger"),
		Period:            new(30),
		RequestMaxBytes:   new(0),
		RequestMaxEntries: new(0),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/1"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.HTTPS{
		ContentType:       new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(5),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("POST"),
		Name:              new("httpslogger2"),
		Period:            new(60),
		RequestMaxBytes:   new(1000),
		RequestMaxEntries: new(0),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/2"),
		ProcessingRegion:  new("none"),
	}

	log3 := gofastly.HTTPS{
		ContentType:       new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("PUT"),
		Name:              new("httpslogger3"),
		Period:            new(120),
		RequestMaxBytes:   new(0),
		RequestMaxEntries: new(0),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/3"),
		ProcessingRegion:  new("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			// Step 1: Basic config
			{
				Config: testAccServiceVCLHTTPSConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "1"),
				),
			},

			// Step 2: Update with response_condition
			{
				Config: testAccServiceVCLHTTPSWithCondition(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&log1WithCondition}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "1"),
				),
			},

			// Step 3: Update with two loggers
			{
				Config: testAccServiceVCLHTTPSConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "2"),
				),
			},

			// Step 4: Compression not specified
			{
				Config: testAccServiceVCLHTTPSConfigCompressionNotSpecified(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSCompressionNotSpecified(&service, []*gofastly.HTTPS{&log3}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingHTTPS_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	https := gofastly.HTTPS{
		CompressionCodec:  new("zstd"),
		ContentType:       new(""),
		GzipLevel:         new(0),
		HeaderName:        new(""),
		HeaderValue:       new(""),
		JSONFormat:        new("0"),
		MessageType:       new("blank"),
		Method:            new("PUT"),
		Name:              new("httpslogger"),
		Period:            new(300),
		RequestMaxBytes:   new(0),
		RequestMaxEntries: new(0),
		ServiceVersion:    new(1),
		TLSHostname:       new(""),
		URL:               new("https://example.com/logs/1"),
		ProcessingRegion:  new("us"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHTTPSComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&https}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_https.#", "1"),
				),
			},
		},
	})
}

// httpsCheckFunc defines a function signature for custom checks.
type httpsCheckFunc func(h, hl *gofastly.HTTPS) error

// testAccCheckFastlyServiceVCLHTTPSAttributesGeneric is a generic helper for checking HTTPS logging attributes.
func testAccCheckFastlyServiceVCLHTTPSAttributesGeneric(service *gofastly.ServiceDetail, https []*gofastly.HTTPS, check httpsCheckFunc) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		httpsList, err := conn.ListHTTPS(context.TODO(), &gofastly.ListHTTPSInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up HTTPS Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(httpsList) != len(https) {
			return fmt.Errorf("https List count mismatch, expected (%d), got (%d)", len(https), len(httpsList))
		}

		log.Printf("[DEBUG] httpsList = %#v\n", httpsList)

		var found int
		for _, h := range https {
			for _, hl := range httpsList {
				if gofastly.ToValue(h.Name) == gofastly.ToValue(hl.Name) {
					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ServiceID
					h.ServiceVersion = service.ActiveVersion.Number

					// Run custom checks if provided.
					if check != nil {
						if err := check(h, hl); err != nil {
							return err
						}
					}

					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					hl.CreatedAt = nil
					hl.UpdatedAt = nil
					if !reflect.DeepEqual(h, hl) {
						return fmt.Errorf("bad match HTTPS logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", h, hl)
					}
					found++
				}
			}
		}

		if found != len(https) {
			return fmt.Errorf("error matching HTTPS Logging rules")
		}

		return nil
	}
}

func testAccCheckFastlyServiceVCLHTTPSAttributes(service *gofastly.ServiceDetail, https []*gofastly.HTTPS, serviceType string) resource.TestCheckFunc {
	return testAccCheckFastlyServiceVCLHTTPSAttributesGeneric(service, https, func(h, hl *gofastly.HTTPS) error {
		// Ignore VCL attributes for Compute and set to whatever is returned from the API.
		if serviceType == ServiceTypeCompute {
			h.Placement = hl.Placement
			h.Format = hl.Format
			h.FormatVersion = hl.FormatVersion
			h.ResponseCondition = hl.ResponseCondition
		}
		return nil
	})
}

func testAccCheckFastlyServiceVCLHTTPSCompressionNotSpecified(service *gofastly.ServiceDetail, https []*gofastly.HTTPS) resource.TestCheckFunc {
	return testAccCheckFastlyServiceVCLHTTPSAttributesGeneric(service, https, func(h, hl *gofastly.HTTPS) error {
		if gofastly.ToValue(hl.GzipLevel) != 0 {
			return fmt.Errorf("Wrong GzipLevel, expected (%d), got (%d)", gofastly.ToValue(h.GzipLevel), gofastly.ToValue(hl.GzipLevel))
		}
		h.GzipLevel = hl.GzipLevel
		return nil
	})
}

func testAccServiceVCLHTTPSConfigCompressionNotSpecified(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-https-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_https {
		name               = "httpslogger3"
		method             = "PUT"
		format             = %q
		period             = 120
		url                = "https://example.com/logs/3"
	}

	force_destroy = true
}
`, name, domain, LoggingFormatUpdate)
}

func testAccServiceVCLHTTPSConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-https-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_https {
		name               = "httpslogger"
		method             = "PUT"
		period			   = 10
		url                = "https://example.com/logs/1"
		compression_codec = "zstd"
    processing_region = "us"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHTTPSComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-https-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_https {
		name               = "httpslogger"
		method             = "PUT"
		url                = "https://example.com/logs/1"
		compression_codec = "zstd"
		period  		  = 300
        processing_region = "us"
	}

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHTTPSConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
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

	logging_https {
		name               = "httpslogger"
		format             = %q
		method             = "POST"
		period			   = 30
		url                = "https://example.com/logs/1"
		compression_codec = "snappy"
	}

	logging_https {
		name               = "httpslogger2"
		format             = %q
		method             = "POST"
		url                = "https://example.com/logs/2"
		period			   = 60
		request_max_bytes  = 1000
		gzip_level				 = 5
	}
	force_destroy = true
}`, name, domain, format, format)
}

func testAccServiceVCLHTTPSWithCondition(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-https-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_https {
    name               = "httpslogger"
    format             = %q
    method             = "POST"
    period             = 30
    url                = "https://example.com/logs/1"
    compression_codec  = "snappy"
    response_condition = "response_condition_test"
    processing_region  = "none"
  }

  condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    statement = "resp.status == 418"
    priority  = 10
  }

  force_destroy = true
}
`, name, domain, format)
}

func TestResourceFastlyFlattenHTTPS(t *testing.T) {
	cases := []struct {
		remote []*gofastly.HTTPS
		local  []map[string]any
	}{
		{
			remote: []*gofastly.HTTPS{
				{
					ServiceVersion:    new(1),
					Name:              new("https-endpoint"),
					URL:               new("https://example.com/logs"),
					RequestMaxEntries: new(10),
					RequestMaxBytes:   new(10),
					CompressionCodec:  new("zstd"),
					ContentType:       new("application/json"),
					MessageType:       new("blank"),
					GzipLevel:         new(0),
					Format:            new(LoggingHTTPSDefaultFormat),
					FormatVersion:     new(2),
					Period:            new(5),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":                "https-endpoint",
					"url":                 "https://example.com/logs",
					"request_max_entries": 10,
					"request_max_bytes":   10,
					"compression_codec":   "zstd",
					"content_type":        "application/json",
					"message_type":        "blank",
					"format":              LoggingHTTPSDefaultFormat,
					"format_version":      2,
					"gzip_level":          0,
					"period":              5,
					"processing_region":   "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHTTPS(c.remote, nil)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}
