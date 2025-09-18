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

func TestAccFastlyServiceVCL_logging_newrelicotlp_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.NewRelicOTLP{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("token"),
		Region:         gofastly.ToPointer("US"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer(LoggingNewRelicOLTPDefaultFormat),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
		URL:               gofastly.ToPointer(""),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.NewRelicOTLP{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("t0k3n"),
		Region:         gofastly.ToPointer("EU"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer(LoggingFormatUpdate),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
		URL:               gofastly.ToPointer(""),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.NewRelicOTLP{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("another-newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("another-token"),
		Region:         gofastly.ToPointer("US"),
		URL:            gofastly.ToPointer("https://example.nr-data.net"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer(LoggingFormatUpdate),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
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
				Config: testAccServiceVCLNewRelicOTLPConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(&service, []*gofastly.NewRelicOTLP{&log1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelicotlp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLNewRelicOTLPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(&service, []*gofastly.NewRelicOTLP{&log1AfterUpdate, &log2}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelicotlp.#", "2"),
				),
				PreventDiskCleanup: true,
			},
		},
	})
}

func testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(service *gofastly.ServiceDetail, newrelicotlp []*gofastly.NewRelicOTLP, serviceType ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		newrelicotlpList, err := conn.ListNewRelicOTLP(context.TODO(), &gofastly.ListNewRelicOTLPInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up NewRelic OTLP Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(newrelicotlpList) != len(newrelicotlp) {
			return fmt.Errorf("newRelic List count mismatch, expected (%d), got (%d)", len(newrelicotlp), len(newrelicotlpList))
		}

		log.Printf("[DEBUG] newrelicotlpList = %#v\n", newrelicotlpList)

		var found int
		for _, d := range newrelicotlp {
			for _, dl := range newrelicotlpList {
				if gofastly.ToValue(d.Name) == gofastly.ToValue(dl.Name) {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ServiceID
					d.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					dl.CreatedAt = nil
					dl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if len(serviceType) > 0 && serviceType[0] == ServiceTypeCompute {
						dl.FormatVersion = d.FormatVersion
						dl.Format = d.Format
						dl.ResponseCondition = d.ResponseCondition
						dl.Placement = d.Placement
						dl.URL = d.URL
					}

					if diff := cmp.Diff(d, dl); diff != "" {
						return fmt.Errorf("bad match NewRelic OTLP logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(newrelicotlp) {
			return fmt.Errorf("error matching NewRelic OTLP Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLNewRelicOTLPConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelicotlp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelicotlp {
    name   = "newrelicotlp-endpoint"
    token  = "token"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLNewRelicOTLPConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelicotlp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelicotlp {
    name   = "newrelicotlp-endpoint"
    token  = "t0k3n"
    format = %q
    region = "EU"
}

  logging_newrelicotlp {
    name   = "another-newrelicotlp-endpoint"
    token  = "another-token"
	url    = "https://example.nr-data.net"
	format = %q
  }

  force_destroy = true
}
`, name, domain, format, format)
}

func TestAccFastlyServiceVCL_logging_newrelicotlp_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.NewRelicOTLP{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("newrelicotlp-endpoint"),
		Token:            gofastly.ToPointer("token"),
		Region:           gofastly.ToPointer(""),
		Placement:        gofastly.ToPointer("none"),
		ProcessingRegion: gofastly.ToPointer("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLNewRelicOTLPComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(&service, []*gofastly.NewRelicOTLP{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_newrelicotlp.#", "1"),
				),
			},
		},
	})
}

func testAccServiceVCLNewRelicOTLPComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelicotlp-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelicotlp {
    name   = "newrelicotlp-endpoint"
    token  = "token"
	placement = "none"
	format_version = "2"
	

  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domain)
}

func TestResourceFastlyFlattenNewRelicOTLP(t *testing.T) {
	var placement, responseCondition, loggingURL *string
	cases := []struct {
		remote []*gofastly.NewRelicOTLP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.NewRelicOTLP{
				{
					ServiceVersion:   gofastly.ToPointer(1),
					Name:             gofastly.ToPointer("newrelicotlp-endpoint"),
					Token:            gofastly.ToPointer("token"),
					Region:           gofastly.ToPointer("US"),
					FormatVersion:    gofastly.ToPointer(2),
					Format:           gofastly.ToPointer(LoggingNewRelicOLTPDefaultFormat),
					ProcessingRegion: gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"format":             gofastly.ToPointer(LoggingNewRelicOLTPDefaultFormat),
					"format_version":     gofastly.ToPointer(2),
					"name":               gofastly.ToPointer("newrelicotlp-endpoint"),
					"placement":          placement, // implies nil
					"region":             gofastly.ToPointer("US"),
					"response_condition": responseCondition, // implies nil
					"token":              gofastly.ToPointer("token"),
					"url":                loggingURL, // implies nil
					"processing_region":  gofastly.ToPointer("eu"),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenNewRelicOTLP(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
