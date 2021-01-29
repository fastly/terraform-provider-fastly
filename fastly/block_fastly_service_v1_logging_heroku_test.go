package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenHeroku(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Heroku
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Heroku{
				{
					ServiceVersion:    1,
					Name:              "heroku-endpoint",
					URL:               "https://example.com",
					Token:             "token",
					Placement:         "none",
					ResponseCondition: "always",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "heroku-endpoint",
					"token":              "token",
					"url":                "https://example.com",
					"placement":          "none",
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"response_condition": "always",
					"format_version":     uint(2),
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

func TestAccFastlyServiceV1_logging_heroku_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		ServiceVersion: 1,
		Name:           "heroku-endpoint",
		URL:            "https://example.com",
		Token:          "s3cr3t",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	log1_after_update := gofastly.Heroku{
		ServiceVersion:    1,
		Name:              "heroku-endpoint",
		URL:               "https://example.com",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		Token:             "secret",
		FormatVersion:     2,
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.Heroku{
		ServiceVersion: 1,
		Name:           "another-heroku-endpoint",
		URL:            "https://new.example.com",
		Token:          "another-token",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1HerokuConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1HerokuAttributes(&service, []*gofastly.Heroku{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_heroku.#", "1"),
				),
			},

			{
				Config: testAccServiceV1HerokuConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1HerokuAttributes(&service, []*gofastly.Heroku{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_heroku.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_heroku_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		ServiceVersion: 1,
		Name:           "heroku-endpoint",
		URL:            "https://example.com",
		Token:          "s3cr3t",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1HerokuComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1HerokuAttributes(&service, []*gofastly.Heroku{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_heroku.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1HerokuAttributes(service *gofastly.ServiceDetail, heroku []*gofastly.Heroku, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		herokuList, err := conn.ListHerokus(&gofastly.ListHerokusInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Heroku Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(herokuList) != len(heroku) {
			return fmt.Errorf("Heroku List count mismatch, expected (%d), got (%d)", len(heroku), len(herokuList))
		}

		log.Printf("[DEBUG] herokuList = %#v\n", herokuList)

		for _, e := range heroku {
			for _, el := range herokuList {
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
						return fmt.Errorf("Bad match Heroku logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceV1HerokuConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1HerokuConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    format             = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
		response_condition = "response_condition_test"
  }

  logging_heroku {
    name   = "another-heroku-endpoint"
    token  = "another-token"
    url    = "https://new.example.com"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1HerokuComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}

  force_destroy = true
}
`, name, domain)
}
