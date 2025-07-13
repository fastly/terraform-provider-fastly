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
					ContentType:       gofastly.ToPointer("application/json"),
					MessageType:       gofastly.ToPointer("blank"),
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
					"content_type":        "application/json",
					"message_type":        "blank",
					"format_version":      2,
					"processing_region":   "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHTTPS(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_httpslogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.HTTPS{
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer("%a %l %u %t %m %U%q %H %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
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
		ContentType:       gofastly.ToPointer(""),
		Format:            gofastly.ToPointer("%a %l %u %t %m %U%q %H %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
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
		Format:            gofastly.ToPointer("%a %l %u %t %m %U%q %H %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
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
		},
	})
}

func TestAccFastlyServiceVCL_httpslogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	https := gofastly.HTTPS{
		ContentType:       gofastly.ToPointer(""),
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
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		method             = "PUT"
		url                = "https://example.com/logs/1"
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
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b"
		method             = "POST"
		url                = "https://example.com/logs/1"
	}

	logging_https {
		name               = "httpslogger2"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		method             = "POST"
		url                = "https://example.com/logs/2"
		request_max_bytes  = 1000
	}
	force_destroy = true
}`, name, domain)
}
