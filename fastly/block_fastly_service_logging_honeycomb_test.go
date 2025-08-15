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

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestAccFastlyServiceVCL_logging_honeycomb_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("dataset"),
		Format:            gofastly.ToPointer(LoggingHoneycombDefaultFormat),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("honeycomb-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("s3cr3t"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("new-dataset"),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("honeycomb-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("secret"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("another-dataset"),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-honeycomb-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("another-token"),
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

func TestAccFastlyServiceVCL_logging_honeycomb_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Honeycomb{
		Dataset:          gofastly.ToPointer("dataset"),
		Name:             gofastly.ToPointer("honeycomb-endpoint"),
		ServiceVersion:   gofastly.ToPointer(1),
		Token:            gofastly.ToPointer("s3cr3t"),
		ProcessingRegion: gofastly.ToPointer("us"),
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
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("honeycomb-endpoint"),
					Token:             gofastly.ToPointer("token"),
					Dataset:           gofastly.ToPointer("dataset"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					Format:            gofastly.ToPointer(LoggingHoneycombDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
					ProcessingRegion:  gofastly.ToPointer("eu"),
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
