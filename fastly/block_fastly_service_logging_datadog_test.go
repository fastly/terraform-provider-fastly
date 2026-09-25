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

func TestAccFastlyServiceLoggingDatadog_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		Format:            new(LoggingDatadogDefaultFormat),
		FormatVersion:     new(2),
		Name:              new("datadog-endpoint"),
		Region:            new("US"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		Token:             new("token"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Datadog{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("datadog-endpoint"),
		Region:            new("EU"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		Token:             new("t0k3n"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Datadog{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("another-datadog-endpoint"),
		Region:            new("US"),
		ResponseCondition: new(""),
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
				Config: testAccServiceVCLDatadogConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_datadog.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLDatadogConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_datadog.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingDatadog_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		ServiceVersion:   new(1),
		Name:             new("datadog-endpoint"),
		Token:            new("token"),
		Region:           new("US"),
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
				Config: testAccServiceVCLDatadogComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_datadog.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLDatadogAttributes(service *gofastly.ServiceDetail, datadog []*gofastly.Datadog, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		datadogList, err := conn.ListDatadog(context.TODO(), &gofastly.ListDatadogInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Datadog Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(datadogList) != len(datadog) {
			return fmt.Errorf("datadog List count mismatch, expected (%d), got (%d)", len(datadog), len(datadogList))
		}

		log.Printf("[DEBUG] datadogList = %#v\n", datadogList)

		var found int
		for _, d := range datadog {
			for _, dl := range datadogList {
				if gofastly.ToValue(d.Name) == gofastly.ToValue(dl.Name) {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ServiceID
					d.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					dl.CreatedAt = nil
					dl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						dl.FormatVersion = d.FormatVersion
						dl.Format = d.Format
						dl.ResponseCondition = d.ResponseCondition
						dl.Placement = d.Placement
					}

					if diff := cmp.Diff(d, dl); diff != "" {
						return fmt.Errorf("bad match Datadog logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(datadog) {
			return fmt.Errorf("error matching Datadog Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLDatadogConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "token"
    region = "US"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLDatadogConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "t0k3n"
    region = "EU"
    format = %q
  }

  logging_datadog {
    name  = "another-datadog-endpoint"
    token = "another-token"
	format = %q
  }

  force_destroy = true
}
`, name, domain, format, format)
}

func testAccServiceVCLDatadogComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "token"
    region = "US"
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

func TestResourceFastlyFlattenDatadog(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Datadog
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Datadog{
				{
					ServiceVersion:   new(1),
					Name:             new("datadog-endpoint"),
					Token:            new("token"),
					Region:           new("US"),
					FormatVersion:    new(2),
					Format:           new(LoggingDatadogDefaultFormat),
					ProcessingRegion: new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":              "datadog-endpoint",
					"token":             "token",
					"region":            "US",
					"format_version":    2,
					"format":            LoggingDatadogDefaultFormat,
					"processing_region": "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDatadog(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
