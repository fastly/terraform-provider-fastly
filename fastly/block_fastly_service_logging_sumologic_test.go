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

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestAccFastlyServiceVCL_logging_sumologic_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Sumologic{
		Format:            gofastly.ToPointer(LoggingSumologicDefaultFormat),
		FormatVersion:     gofastly.ToPointer(2),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("sumologic-endpoint"),
		ResponseCondition: gofastly.ToPointer("test_response_condition"),
		ServiceVersion:    gofastly.ToPointer(1),
		URL:               gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.Sumologic{
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("sumologic-endpoint"),
		ResponseCondition: gofastly.ToPointer("test_response_condition"),
		ServiceVersion:    gofastly.ToPointer(1),
		URL:               gofastly.ToPointer("https://collectors.sumologic.com/receiver/2"),
		ProcessingRegion:  gofastly.ToPointer("eu"),
	}

	log2 := gofastly.Sumologic{
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("another-sumologic-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		URL:               gofastly.ToPointer("https://collectors.sumologic.com/receiver/3"),
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
				Config: testAccServiceVCLSumologicConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSumologicAttributes(&service, []*gofastly.Sumologic{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_sumologic.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLSumologicConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSumologicAttributes(&service, []*gofastly.Sumologic{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_sumologic.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_sumologic_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Sumologic{
		Name:             gofastly.ToPointer("sumologic-endpoint"),
		ServiceVersion:   gofastly.ToPointer(1),
		URL:              gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
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
				Config: testAccServiceVCLSumologicComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLSumologicAttributes(&service, []*gofastly.Sumologic{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_sumologic.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLSumologicAttributes(service *gofastly.ServiceDetail, sumologic []*gofastly.Sumologic, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		sumologicList, err := conn.ListSumologics(context.TODO(), &gofastly.ListSumologicsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up SumoLogic Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(sumologicList) != len(sumologic) {
			return fmt.Errorf("sumologic List count mismatch, expected (%d), got (%d)", len(sumologic), len(sumologicList))
		}

		log.Printf("[DEBUG] sumologicList = %#v\n", sumologicList)

		var found int
		for _, s := range sumologic {
			for _, sl := range sumologicList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(sl.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ServiceID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					sl.CreatedAt = nil
					sl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						sl.FormatVersion = s.FormatVersion
						sl.Format = s.Format
						sl.ResponseCondition = s.ResponseCondition
						sl.Placement = s.Placement
						sl.MessageType = s.MessageType
					}

					if diff := cmp.Diff(s, sl); diff != "" {
						return fmt.Errorf("bad match SumoLogic logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(sumologic) {
			return fmt.Errorf("error matching SumoLogic Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLSumologicComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-sumologic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_sumologic {
    name = "sumologic-endpoint"
    url = "https://collectors.sumologic.com/receiver/1"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSumologicConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-sumologic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  condition {
    name      = "test_response_condition"
    type      = "RESPONSE"
    priority  = 5
    statement = "resp.status >= 400 && resp.status < 600"
  }

  logging_sumologic {
    name = "sumologic-endpoint"
    url = "https://collectors.sumologic.com/receiver/1"
    message_type = "classic"
    response_condition = "test_response_condition"
    processing_region = "us"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSumologicConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-sumologic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  condition {
    name      = "test_response_condition"
    type      = "RESPONSE"
    priority  = 5
    statement = "resp.status >= 400 && resp.status < 600"
  }

  logging_sumologic {
    name = "sumologic-endpoint"
    url = "https://collectors.sumologic.com/receiver/2"
    format = %q
    message_type = "blank"
    response_condition = "test_response_condition"
    processing_region = "eu"
  }

  logging_sumologic {
    name = "another-sumologic-endpoint"
    url = "https://collectors.sumologic.com/receiver/3"
    format = %q
    message_type = "classic"
  }

  force_destroy = true
}`, name, domain, format, format)
}

func TestResourceFastlyFlattenSumologic(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Sumologic
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Sumologic{
				{
					Name:              gofastly.ToPointer("sumo collector"),
					URL:               gofastly.ToPointer("https://collectors.sumologic.com/receiver/1"),
					Format:            gofastly.ToPointer(LoggingSumologicDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
					MessageType:       gofastly.ToPointer("classic"),
					ResponseCondition: gofastly.ToPointer("condition 1"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "sumo collector",
					"url":                "https://collectors.sumologic.com/receiver/1",
					"format":             LoggingSumologicDefaultFormat,
					"format_version":     2,
					"message_type":       "classic",
					"response_condition": "condition 1",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSumologics(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
