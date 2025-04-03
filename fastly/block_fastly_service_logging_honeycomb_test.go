package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var honeycombDefaultFormat = `{
  "time":"%{begin:%Y-%m-%dT%H:%M:%SZ}t",
  "data":  {
    "service_id":"%{req.service_id}V",
    "time_elapsed":%D,
    "request":"%m",
    "host":"%{Fastly-Orig-Host}i",
    "url":"%{cstr_escape(req.url)}V",
    "protocol":"%H",
    "is_ipv6":%{if(req.is_ipv6, "true", "false")}V,
    "is_tls":%{if(req.is_ssl, "true", "false")}V,
    "is_h2":%{if(fastly_info.is_h2, "true", "false")}V,
    "client_ip":"%h",
    "geo_city":"%{client.geo.city.utf8}V",
    "geo_country_code":"%{client.geo.country_code}V",
    "server_datacenter":"%{server.datacenter}V",
    "request_referer":"%{Referer}i",
    "request_user_agent":"%{User-Agent}i",
    "request_accept_content":"%{Accept}i",
    "request_accept_language":"%{Accept-Language}i",
    "request_accept_charset":"%{Accept-Charset}i",
    "cache_status":"%{regsub(fastly_info.state, "^(HIT-(SYNTH)|(HITPASS|HIT|MISS|PASS|ERROR|PIPE)).*", "\\2\\3") }V",
    "status":"%s",
    "content_type":"%{Content-Type}o",
    "req_header_size":%{req.header_bytes_read}V,
    "req_body_size":%{req.body_bytes_read}V,
    "resp_header_size":%{resp.header_bytes_written}V,
    "resp_body_size":%{resp.body_bytes_written}V
  }
}`

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
					Format:            gofastly.ToPointer(honeycombDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "honeycomb-endpoint",
					"token":              "token",
					"dataset":            "dataset",
					"placement":          "none",
					"response_condition": "always",
					"format":             honeycombDefaultFormat,
					"format_version":     2,
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

func TestAccFastlyServiceVCL_logging_honeycomb_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("dataset"),
		Format:            gofastly.ToPointer(appendNewLine(honeycombDefaultFormat)),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("honeycomb-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("s3cr3t"),
	}

	log1AfterUpdate := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("new-dataset"),
		Format:            gofastly.ToPointer(appendNewLine(honeycombDefaultFormat)),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("honeycomb-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("secret"),
	}

	log2 := gofastly.Honeycomb{
		Dataset:           gofastly.ToPointer("another-dataset"),
		Format:            gofastly.ToPointer(appendNewLine(honeycombDefaultFormat)),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-honeycomb-endpoint"),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
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
		Dataset:        gofastly.ToPointer("dataset"),
		Name:           gofastly.ToPointer("honeycomb-endpoint"),
		ServiceVersion: gofastly.ToPointer(1),
		Token:          gofastly.ToPointer("s3cr3t"),
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
		honeycombList, err := conn.ListHoneycombs(&gofastly.ListHoneycombsInput{
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
    format = <<EOF
`+escapePercentSign(honeycombDefaultFormat)+`
EOF
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLHoneycombConfigUpdate(name, domain string) string {
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
    format = <<EOF
`+escapePercentSign(honeycombDefaultFormat)+`
EOF
    response_condition = "response_condition_test"
		placement = "none"
  }

  logging_honeycomb {
    name   = "another-honeycomb-endpoint"
    token  = "another-token"
		dataset = "another-dataset"
    format = <<EOF
`+escapePercentSign(honeycombDefaultFormat)+`
EOF
    response_condition = "response_condition_test"
		placement = "none"
  }

  force_destroy = true
}
`, name, domain)
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
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
