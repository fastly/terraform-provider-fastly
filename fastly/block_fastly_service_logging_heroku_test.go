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

func TestAccFastlyServiceLoggingHeroku_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		Format:            new(LoggingHerokuDefaultFormat),
		FormatVersion:     new(2),
		Name:              new("heroku-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		Token:             new("s3cr3t"),
		URL:               new("https://example.com"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Heroku{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("heroku-endpoint"),
		Placement:         new("none"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		Token:             new("secret"),
		URL:               new("https://example.com"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Heroku{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("another-heroku-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		Token:             new("another-token"),
		URL:               new("https://new.example.com"),
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
				Config: testAccServiceVCLHerokuConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHerokuAttributes(&service, []*gofastly.Heroku{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_heroku.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLHerokuConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHerokuAttributes(&service, []*gofastly.Heroku{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_heroku.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingHeroku_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		Name:             new("heroku-endpoint"),
		ServiceVersion:   new(1),
		Token:            new("s3cr3t"),
		URL:              new("https://example.com"),
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
				Config: testAccServiceVCLHerokuComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLHerokuAttributes(&service, []*gofastly.Heroku{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_heroku.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLHerokuAttributes(service *gofastly.ServiceDetail, heroku []*gofastly.Heroku, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		herokuList, err := conn.ListHerokus(context.TODO(), &gofastly.ListHerokusInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Heroku Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(herokuList) != len(heroku) {
			return fmt.Errorf("heroku List count mismatch, expected (%d), got (%d)", len(heroku), len(herokuList))
		}

		log.Printf("[DEBUG] herokuList = %#v\n", herokuList)

		for _, e := range heroku {
			for _, el := range herokuList {
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
						return fmt.Errorf("bad match Heroku logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLHerokuConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-heroku-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_heroku {
    name   = "heroku-endpoint"
    token  = "s3cr3t"
	url    = "https://example.com"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHerokuConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-heroku-logging"
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

  logging_heroku {
    name               = "heroku-endpoint"
    url                = "https://example.com"
    placement          = "none"
    token              = "secret"
    format             = %q
	response_condition = "response_condition_test"
  }

  logging_heroku {
    name   = "another-heroku-endpoint"
    token  = "another-token"
    url    = "https://new.example.com"
    format = %q
  }

  force_destroy = true
}
`, name, domain, format, format)
}

func testAccServiceVCLHerokuComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-heroku-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_heroku {
    name   = "heroku-endpoint"
    token  = "s3cr3t"
    url    = "https://example.com"
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

func TestResourceFastlyFlattenHeroku(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Heroku
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Heroku{
				{
					ServiceVersion:    new(1),
					Name:              new("heroku-endpoint"),
					URL:               new("https://example.com"),
					Token:             new("token"),
					Placement:         new("none"),
					ResponseCondition: new("always"),
					Format:            new(LoggingHerokuDefaultFormat),
					FormatVersion:     new(2),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "heroku-endpoint",
					"token":              "token",
					"url":                "https://example.com",
					"placement":          "none",
					"format":             LoggingHerokuDefaultFormat,
					"response_condition": "always",
					"format_version":     2,
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHeroku(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
