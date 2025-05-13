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

func TestResourceFastlyFlattenNewRelic(t *testing.T) {
	cases := []struct {
		remote []*gofastly.NewRelic
		local  []map[string]any
	}{
		{
			remote: []*gofastly.NewRelic{
				{
					ServiceVersion: gofastly.ToPointer(1),
					Name:           gofastly.ToPointer("newrelic-endpoint"),
					Token:          gofastly.ToPointer("token"),
					Region:         gofastly.ToPointer("US"),
					FormatVersion:  gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":           "newrelic-endpoint",
					"token":          "token",
					"region":         "US",
					"format_version": 2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenNewRelic(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

var newrelicDefaultFormat = `{
  "time_elapsed":%{time.elapsed.usec}V,
  "is_tls":%{if(req.is_ssl, "true", "false")}V,
  "client_ip":"%{req.http.Fastly-Client-IP}V",
  "geo_city":"%{client.geo.city}V",
  "geo_country_code":"%{client.geo.country_code}V",
  "request":"%{req.request}V",
  "host":"%{req.http.Fastly-Orig-Host}V",
  "url":"%{json.escape(req.url)}V",
  "request_referer":"%{json.escape(req.http.Referer)}V",
  "request_user_agent":"%{json.escape(req.http.User-Agent)}V",
  "request_accept_language":"%{json.escape(req.http.Accept-Language)}V",
  "request_accept_charset":"%{json.escape(req.http.Accept-Charset)}V",
  "cache_status":"%{regsub(fastly_info.state, "^(HIT-(SYNTH)|(HITPASS|HIT|MISS|PASS|ERROR|PIPE)).*", "\2\3") }V"
}`

func TestAccFastlyServiceVCL_logging_newrelic_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.NewRelic{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("newrelic-endpoint"),
		Region:            gofastly.ToPointer("US"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("token"),
	}

	log1AfterUpdate := gofastly.NewRelic{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("newrelic-endpoint"),
		Region:            gofastly.ToPointer("EU"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("t0k3n"),
	}

	log2 := gofastly.NewRelic{
		Format:            gofastly.ToPointer(appendNewLine(newrelicDefaultFormat)),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-newrelic-endpoint"),
		Region:            gofastly.ToPointer("US"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("another-token"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLNewRelicConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicAttributes(&service, []*gofastly.NewRelic{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelic.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLNewRelicConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicAttributes(&service, []*gofastly.NewRelic{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelic.#", "2"),
				),
				PreventDiskCleanup: true,
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_newrelic_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.NewRelic{
		Format:         gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:  gofastly.ToPointer(2),
		Name:           gofastly.ToPointer("newrelic-endpoint"),
		Region:         gofastly.ToPointer("US"),
		ServiceVersion: gofastly.ToPointer(1),
		Token:          gofastly.ToPointer("token"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLNewRelicComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicAttributes(&service, []*gofastly.NewRelic{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_newrelic.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLNewRelicAttributes(service *gofastly.ServiceDetail, newrelic []*gofastly.NewRelic, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		newrelicList, err := conn.ListNewRelic(&gofastly.ListNewRelicInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up NewRelic Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(newrelicList) != len(newrelic) {
			return fmt.Errorf("newRelic List count mismatch, expected (%d), got (%d)", len(newrelic), len(newrelicList))
		}

		log.Printf("[DEBUG] newrelicList = %#v\n", newrelicList)

		var found int
		for _, d := range newrelic {
			for _, dl := range newrelicList {
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
						return fmt.Errorf("bad match NewRelic logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(newrelic) {
			return fmt.Errorf("error matching NewRelic Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLNewRelicComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelic {
    name   = "newrelic-endpoint"
    token  = "token"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLNewRelicConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelic {
    name   = "newrelic-endpoint"
    token  = "token"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLNewRelicConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-newrelic-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_newrelic {
    name   = "newrelic-endpoint"
    token  = "t0k3n"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    region = "EU"
  }

  logging_newrelic {
    name  = "another-newrelic-endpoint"
    token = "another-token"
		format = <<EOF
`+escapePercentSign(newrelicDefaultFormat)+`
EOF
  }

  force_destroy = true
}
`, name, domain)
}
