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

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestAccFastlyServiceVCL_httpslogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.HTTPS{
		CompressionCodec:  gofastly.ToPointer("zstd"),
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer(LoggingHTTPSDefaultFormat),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		HeaderName:        gofastly.ToPointer(""),
		HeaderValue:       gofastly.ToPointer(""),
		JSONFormat:        gofastly.ToPointer("0"),
		MessageType:       gofastly.ToPointer("blank"),
		Method:            gofastly.ToPointer("PUT"),
		Name:              gofastly.ToPointer("httpslogger"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TLSHostname:       gofastly.ToPointer(""),
		URL:               gofastly.ToPointer("https://example.com/logs/1"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.HTTPS{
		CompressionCodec:  gofastly.ToPointer("snappy"),
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		HeaderName:        gofastly.ToPointer(""),
		HeaderValue:       gofastly.ToPointer(""),
		JSONFormat:        gofastly.ToPointer("0"),
		MessageType:       gofastly.ToPointer("blank"),
		Method:            gofastly.ToPointer("POST"),
		Name:              gofastly.ToPointer("httpslogger"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TLSHostname:       gofastly.ToPointer(""),
		URL:               gofastly.ToPointer("https://example.com/logs/1"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.HTTPS{
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(5),
		HeaderName:        gofastly.ToPointer(""),
		HeaderValue:       gofastly.ToPointer(""),
		JSONFormat:        gofastly.ToPointer("0"),
		MessageType:       gofastly.ToPointer("blank"),
		Method:            gofastly.ToPointer("POST"),
		Name:              gofastly.ToPointer("httpslogger2"),
		RequestMaxBytes:   gofastly.ToPointer(1000),
		RequestMaxEntries: gofastly.ToPointer(0),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TLSHostname:       gofastly.ToPointer(""),
		URL:               gofastly.ToPointer("https://example.com/logs/2"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log3 := gofastly.HTTPS{
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		HeaderName:        gofastly.ToPointer(""),
		HeaderValue:       gofastly.ToPointer(""),
		JSONFormat:        gofastly.ToPointer("0"),
		MessageType:       gofastly.ToPointer("blank"),
		Method:            gofastly.ToPointer("PUT"),
		Name:              gofastly.ToPointer("httpslogger3"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		TLSHostname:       gofastly.ToPointer(""),
		URL:               gofastly.ToPointer("https://example.com/logs/3"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHTTPSConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLHTTPSConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSAttributes(&service, []*gofastly.HTTPS{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLHTTPSConfigCompressionNotSpecified(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHTTPSCompressionNotSpecified(&service, []*gofastly.HTTPS{&log3}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_https.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_httpslogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	https := gofastly.HTTPS{
		CompressionCodec:  gofastly.ToPointer("zstd"),
		ContentType:       gofastly.ToPointer(""),
		GzipLevel:         gofastly.ToPointer(0),
		HeaderName:        gofastly.ToPointer(""),
		HeaderValue:       gofastly.ToPointer(""),
		JSONFormat:        gofastly.ToPointer("0"),
		MessageType:       gofastly.ToPointer("blank"),
		Method:            gofastly.ToPointer("PUT"),
		Name:              gofastly.ToPointer("httpslogger"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		ServiceVersion:    gofastly.ToPointer(1),
		TLSHostname:       gofastly.ToPointer(""),
		URL:               gofastly.ToPointer("https://example.com/logs/1"),
		ProcessingRegion:  gofastly.ToPointer("us"),
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

func testAccCheckFastlyServiceVCLHTTPSAttributes(service *gofastly.ServiceDetail, https []*gofastly.HTTPS, serviceType string) resource.TestCheckFunc {
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

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						h.Placement = hl.Placement
						h.Format = hl.Format
						h.FormatVersion = hl.FormatVersion
						h.ResponseCondition = hl.ResponseCondition
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

func testAccCheckFastlyServiceVCLHTTPSCompressionNotSpecified(service *gofastly.ServiceDetail, https []*gofastly.HTTPS, serviceType string) resource.TestCheckFunc {
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
					if gofastly.ToValue(hl.GzipLevel) != 0 {
						return fmt.Errorf("Wrong GzipLevel, expected (%d), got (%d)", gofastly.ToValue(h.GzipLevel), gofastly.ToValue(hl.GzipLevel))
					}
					h.GzipLevel = hl.GzipLevel

					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ServiceID
					h.ServiceVersion = service.ActiveVersion.Number

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
		url                = "https://example.com/logs/1"
		compression_codec = "snappy"
	}

	logging_https {
		name               = "httpslogger2"
		format             = %q
		method             = "POST"
		url                = "https://example.com/logs/2"
		request_max_bytes  = 1000
		gzip_level				 = 5
	}
	force_destroy = true
}`, name, domain, format, format)
}

func TestResourceFastlyFlattenHTTPS(t *testing.T) {
	cases := []struct {
		remote []*gofastly.HTTPS
		local  []map[string]any
	}{
		{
			remote: []*gofastly.HTTPS{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("https-endpoint"),
					URL:               gofastly.ToPointer("https://example.com/logs"),
					RequestMaxEntries: gofastly.ToPointer(10),
					RequestMaxBytes:   gofastly.ToPointer(10),
					CompressionCodec:  gofastly.ToPointer("zstd"),
					ContentType:       gofastly.ToPointer("application/json"),
					MessageType:       gofastly.ToPointer("blank"),
					GzipLevel:         gofastly.ToPointer(0),
					Format:            gofastly.ToPointer(LoggingHTTPSDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
					ProcessingRegion:  gofastly.ToPointer("eu"),
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
					"processing_region":   "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHTTPS(c.remote, nil)
		if !reflect.DeepEqual(out, c.local) {
			fmt.Printf("")
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}
