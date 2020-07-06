package fastly

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenHTTPS(t *testing.T) {
	cases := []struct {
		remote []*gofastly.HTTPS
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.HTTPS{
				{
					Version:           1,
					Name:              "https-endpoint",
					URL:               "https://example.com/logs",
					RequestMaxEntries: 10,
					RequestMaxBytes:   10,
					ContentType:       "application/json",
					MessageType:       "blank",
					FormatVersion:     2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":                "https-endpoint",
					"url":                 "https://example.com/logs",
					"request_max_entries": uint(10),
					"request_max_bytes":   uint(10),
					"content_type":        "application/json",
					"message_type":        "blank",
					"format_version":      uint(2),
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

func TestAccFastlyServiceV1_httpslogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	nameCompute := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.HTTPS{
		Version: 1,
		Name:    "httpslogger",
		URL:     "https://example.com/logs/1",
		Method:  "PUT",

		Format:            "%a %l %u %t %m %U%q %H %>s %b %T",
		RequestMaxEntries: 0,
		RequestMaxBytes:   0,
		MessageType:       "blank",
		FormatVersion:     2,
		JSONFormat:        "0",
	}

	logCompute := gofastly.HTTPS{
		Version:           1,
		Name:              "httpslogger",
		URL:               "https://example.com/logs/1",
		Method:            "PUT",
		RequestMaxEntries: 0,
		RequestMaxBytes:   0,
		MessageType:       "blank",
		JSONFormat:        "0",
	}

	log1_after_update := gofastly.HTTPS{
		Version: 1,
		Name:    "httpslogger",
		URL:     "https://example.com/logs/1",
		Method:  "POST",

		Format:            "%a %l %u %t %m %U%q %H %>s %b",
		RequestMaxEntries: 0,
		RequestMaxBytes:   0,
		MessageType:       "blank",
		FormatVersion:     2,
		JSONFormat:        "0",
	}

	log2 := gofastly.HTTPS{
		Version: 1,
		Name:    "httpslogger2",
		URL:     "https://example.com/logs/2",
		Method:  "POST",

		Format:            "%a %l %u %t %m %U%q %H %>s %b %T",
		RequestMaxEntries: 0,
		RequestMaxBytes:   1000,
		MessageType:       "blank",
		FormatVersion:     2,
		JSONFormat:        "0",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1HTTPSConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1HTTPSAttributes(&service, []*gofastly.HTTPS{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "httpslogging.#", "1"),
				),
			},

			{
				Config: testAccServiceV1HTTPSComputeConfig(nameCompute, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1HTTPSAttributes(&service, []*gofastly.HTTPS{&logCompute}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", nameCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "httpslogging.#", "1"),
				),
			},

			{
				Config: testAccServiceV1HTTPSConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1HTTPSAttributes(&service, []*gofastly.HTTPS{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "httpslogging.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1HTTPSAttributes(service *gofastly.ServiceDetail, https []*gofastly.HTTPS, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		httpsList, err := conn.ListHTTPS(&gofastly.ListHTTPSInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up HTTPS Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(httpsList) != len(https) {
			return fmt.Errorf("HTTPS List count mismatch, expected (%d), got (%d)", len(https), len(httpsList))
		}

		log.Printf("[DEBUG] httpsList = %#v\n", httpsList)

		var found int
		for _, h := range https {
			for _, hl := range httpsList {
				if h.Name == hl.Name {
					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ID
					h.Version = service.ActiveVersion.Number

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						h.Placement = hl.Placement
						h.Format = hl.Format
						h.FormatVersion = hl.FormatVersion
						h.ResponseCondition = hl.ResponseCondition
					}

					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					hl.CreatedAt = nil
					hl.UpdatedAt = nil
					if !reflect.DeepEqual(h, hl) {
						return fmt.Errorf("Bad match HTTPS logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", h, hl)
					}
					found++
				}
			}
		}

		if found != len(https) {
			return fmt.Errorf("Error matching HTTPS Logging rules")
		}

		return nil
	}
}

func testAccServiceV1HTTPSConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-https-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	httpslogging {
		name               = "httpslogger"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		method             = "PUT"
		url                = "https://example.com/logs/1"
	}

	force_destroy = true
}
`, name, domain)
}
func testAccServiceV1HTTPSComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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

	httpslogging {
		name               = "httpslogger"
		method             = "PUT"
		url                = "https://example.com/logs/1"
	}

package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1HTTPSConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
	name = "%s"
	domain {
		name    = "%s"
		comment = "tf-testing-domain"
	}
	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	httpslogging {
		name               = "httpslogger"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b"
		method             = "POST"
		url                = "https://example.com/logs/1"
	}

	httpslogging {
		name               = "httpslogger2"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		method             = "POST"
		url                = "https://example.com/logs/2"
		request_max_bytes  = 1000
	}
	force_destroy = true
}`, name, domain)
}
