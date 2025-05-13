package fastly

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

func TestResourceFastlyFlattenLogshuttle(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Logshuttle
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Logshuttle{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("logshuttle-endpoint"),
					Token:             gofastly.ToPointer("token"),
					URL:               gofastly.ToPointer("https://example.com"),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					FormatVersion:     gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "logshuttle-endpoint",
					"token":              "token",
					"url":                "https://example.com",
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
					"placement":          "none",
					"response_condition": "always",
					"format_version":     2,
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

func TestAccFastlyServiceVCL_logging_logshuttle_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Logshuttle{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("logshuttle-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("s3cr3t"),
		URL:               gofastly.ToPointer("https://example.com"),
	}

	log1AfterUpdate := gofastly.Logshuttle{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("logshuttle-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("secret"),
		URL:               gofastly.ToPointer("https://new.example.com"),
	}

	log2 := gofastly.Logshuttle{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-logshuttle-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("another-token"),
		URL:               gofastly.ToPointer("https://another.example.com"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogshuttleConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_logshuttle.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLLogshuttleConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_logshuttle.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_logshuttle_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Logshuttle{
		Name:           gofastly.ToPointer("logshuttle-endpoint"),
		ServiceVersion: gofastly.ToPointer(1),
		Token:          gofastly.ToPointer("s3cr3t"),
		URL:            gofastly.ToPointer("https://example.com"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogshuttleComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLLogshuttleAttributes(&service, []*gofastly.Logshuttle{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_logshuttle.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLLogshuttleAttributes(service *gofastly.ServiceDetail, logshuttle []*gofastly.Logshuttle, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		logshuttleList, err := conn.ListLogshuttles(&gofastly.ListLogshuttlesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Log Shuttle Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(logshuttleList) != len(logshuttle) {
			return fmt.Errorf("log Shuttle List count mismatch, expected (%d), got (%d)", len(logshuttle), len(logshuttleList))
		}

		log.Printf("[DEBUG] logshuttleList = %#v\n", logshuttleList)

		for _, e := range logshuttle {
			for _, el := range logshuttleList {
				if gofastly.ToValue(e.Name) == gofastly.ToValue(el.Name) {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ServiceID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
						return fmt.Errorf("bad match Log Shuttle logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLLogshuttleConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLLogshuttleConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLLogshuttleComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
