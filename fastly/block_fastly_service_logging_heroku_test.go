package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenHeroku(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Heroku
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Heroku{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("heroku-endpoint"),
					URL:               gofastly.ToPointer("https://example.com"),
					Token:             gofastly.ToPointer("token"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					FormatVersion:     gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "heroku-endpoint",
					"token":              "token",
					"url":                "https://example.com",
					"placement":          "none",
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"response_condition": "always",
					"format_version":     2,
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

func TestAccFastlyServiceVCL_logging_heroku_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("heroku-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("s3cr3t"),
		URL:               gofastly.ToPointer("https://example.com"),
	}

	log1AfterUpdate := gofastly.Heroku{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("heroku-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("secret"),
		URL:               gofastly.ToPointer("https://example.com"),
	}

	log2 := gofastly.Heroku{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-heroku-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("another-token"),
		URL:               gofastly.ToPointer("https://new.example.com"),
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

func TestAccFastlyServiceVCL_logging_heroku_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Heroku{
		Name:           gofastly.ToPointer("heroku-endpoint"),
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
		herokuList, err := conn.ListHerokus(&gofastly.ListHerokusInput{
			ServiceID:      gofastly.ToValue(service.ID),
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
					e.ServiceID = service.ID
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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHerokuConfigUpdate(name, domain string) string {
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
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
