package fastly

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func TestAccFastlyServiceLoggingHoneycomb_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Honeycomb{
		Dataset:           new("dataset"),
		Format:            new(LoggingHoneycombDefaultFormat),
		FormatVersion:     new(2),
		Name:              new("honeycomb-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		Token:             new("s3cr3t"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Honeycomb{
		Dataset:           new("new-dataset"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("honeycomb-endpoint"),
		Placement:         new("none"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		Token:             new("secret"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Honeycomb{
		Dataset:           new("another-dataset"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("another-honeycomb-endpoint"),
		Placement:         new("none"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		Token:             new("another-token"),
		ProcessingRegion:  new("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHoneycombConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHoneycombAttributes(&service, []*gofastly.Honeycomb{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_honeycomb.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLHoneycombConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHoneycombAttributes(&service, []*gofastly.Honeycomb{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_honeycomb.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingHoneycomb_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Honeycomb{
		Dataset:          new("dataset"),
		Name:             new("honeycomb-endpoint"),
		ServiceVersion:   new(1),
		Token:            new("s3cr3t"),
		ProcessingRegion: new("us"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHoneycombComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLHoneycombAttributes(&service, []*gofastly.Honeycomb{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_honeycomb.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLHoneycombAttributes(service *gofastly.ServiceDetail, honeycomb []*gofastly.Honeycomb, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		honeycombList, err := conn.ListHoneycombs(context.TODO(), &gofastly.ListHoneycombsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Honeycomb Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(honeycombList) != len(honeycomb) {
			return fmt.Errorf("honeycomb List count mismatch, expected (%d), got (%d)", len(honeycomb), len(honeycombList))
		}

		log.Printf("[DEBUG] honeycombList = %#v\n", honeycombList)

		for _, e := range honeycomb {
			for _, el := range honeycombList {
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
						return fmt.Errorf("bad match Honeycomb logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLHoneycombConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-honeycomb-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_honeycomb {
    name   = "honeycomb-endpoint"
    token  = "s3cr3t"
	dataset = "dataset"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHoneycombConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-honeycomb-logging"
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

  logging_honeycomb {
    name   = "honeycomb-endpoint"
    token  = "secret"
	dataset = "new-dataset"
    format = %q
    response_condition = "response_condition_test"
	placement = "none"
  }

  logging_honeycomb {
    name   = "another-honeycomb-endpoint"
    token  = "another-token"
	dataset = "another-dataset"
    format = %q
    response_condition = "response_condition_test"
	placement = "none"
  }

  force_destroy = true
}
`, name, domain, format, format)
}

func testAccServiceVCLHoneycombComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-honeycomb-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_honeycomb {
    name   = "honeycomb-endpoint"
    token  = "s3cr3t"
    dataset = "dataset"
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

func TestResourceFastlyFlattenHoneycomb(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Honeycomb
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Honeycomb{
				{
					ServiceVersion:    new(1),
					Name:              new("honeycomb-endpoint"),
					Token:             new("token"),
					Dataset:           new("dataset"),
					Placement:         new("none"),
					ResponseCondition: new("always"),
					Format:            new(LoggingHoneycombDefaultFormat),
					FormatVersion:     new(2),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "honeycomb-endpoint",
					"token":              "token",
					"dataset":            "dataset",
					"placement":          "none",
					"response_condition": "always",
					"format":             LoggingHoneycombDefaultFormat,
					"format_version":     2,
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHoneycomb(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
