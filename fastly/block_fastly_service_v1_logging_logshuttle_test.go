package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenLogshuttle(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Logshuttle
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Logshuttle{
				{
					ServiceVersion:    1,
					Name:              "logshuttle-endpoint",
					Token:             "token",
					URL:               "https://example.com",
					Format:            "%h %l %u %t \"%r\" %>s %b %T",
					Placement:         "none",
					ResponseCondition: "always",
					FormatVersion:     2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "logshuttle-endpoint",
					"token":              "token",
					"url":                "https://example.com",
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
					"placement":          "none",
					"response_condition": "always",
					"format_version":     uint(2),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenLogshuttle(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceV1_logging_logshuttle_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Logshuttle{
		ServiceVersion: 1,
		Name:           "logshuttle-endpoint",
		Token:          "s3cr3t",
		URL:            "https://example.com",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	log1_after_update := gofastly.Logshuttle{
		ServiceVersion: 1,
		Name:           "logshuttle-endpoint",
		Token:          "secret",
		URL:            "https://new.example.com",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.Logshuttle{
		ServiceVersion:    1,
		Name:              "another-logshuttle-endpoint",
		Token:             "another-token",
		URL:               "https://another.example.com",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		FormatVersion:     2,
		Format:            "%h %l %u %t \"%r\" %>s %b",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1LogshuttleConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_logshuttle.#", "1"),
				),
			},

			{
				Config: testAccServiceV1LogshuttleConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_logshuttle.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_logshuttle_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Logshuttle{
		ServiceVersion: 1,
		Name:           "logshuttle-endpoint",
		Token:          "s3cr3t",
		URL:            "https://example.com",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1LogshuttleComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1LogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_logshuttle.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1LogshuttleAttributes(service *gofastly.ServiceDetail, logshuttle []*gofastly.Logshuttle, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		logshuttleList, err := conn.ListLogshuttles(&gofastly.ListLogshuttlesInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Log Shuttle Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(logshuttleList) != len(logshuttle) {
			return fmt.Errorf("Log Shuttle List count mismatch, expected (%d), got (%d)", len(logshuttle), len(logshuttleList))
		}

		log.Printf("[DEBUG] logshuttleList = %#v\n", logshuttleList)

		for _, e := range logshuttle {
			for _, el := range logshuttleList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					el.CreatedAt = nil
					el.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						el.FormatVersion = e.FormatVersion
						el.Format = e.Format
						el.ResponseCondition = e.ResponseCondition
						el.Placement = e.Placement
					}

					if diff := cmp.Diff(e, el); diff != "" {
						return fmt.Errorf("Bad match Log Shuttle logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceV1LogshuttleConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-logshuttle-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_logshuttle {
    name   = "logshuttle-endpoint"
    token  = "s3cr3t"
		url    = "https://example.com"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1LogshuttleConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-logshuttle-logging"
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

  logging_logshuttle {
    name   = "logshuttle-endpoint"
    token  = "secret"
    url    = "https://new.example.com"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
  }

  logging_logshuttle {
    name   = "another-logshuttle-endpoint"
    token  = "another-token"
		url    = "https://another.example.com"
    placement = "none"
		response_condition = "response_condition_test"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1LogshuttleComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-logshuttle-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_logshuttle {
    name   = "logshuttle-endpoint"
    token  = "s3cr3t"
	url    = "https://example.com"
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}
