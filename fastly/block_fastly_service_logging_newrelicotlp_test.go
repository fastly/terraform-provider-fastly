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

func TestResourceFastlyFlattenNewRelicOTLP(t *testing.T) {
	cases := []struct {
		remote []*gofastly.NewRelicOTLP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.NewRelicOTLP{
				{
					ServiceVersion: 1,
					Name:           "newrelicotlp-endpoint",
					Token:          "token",
					Region:         "US",
					FormatVersion:  2,
				},
			},
			local: []map[string]any{
				{
					"name":           "newrelicotlp-endpoint",
					"token":          "token",
					"region":         "US",
					"format_version": 2,
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

var newrelicotlpDefaultFormat = `{
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

func TestAccFastlyServiceVCL_logging_newrelicotlp_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.NewRelicOTLP{
		ServiceVersion: 1,
		Name:           "newrelicotlp-endpoint",
		Token:          "token",
		Region:         "US",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	log1AfterUpdate := gofastly.NewRelicOTLP{
		ServiceVersion: 1,
		Name:           "newrelicotlp-endpoint",
		Token:          "t0k3n",
		Region:         "EU",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.NewRelicOTLP{
		ServiceVersion: 1,
		Name:           "another-newrelicotlp-endpoint",
		Token:          "another-token",
		Region:         "US",
		URL:            "https://example.nr-data.net",
		FormatVersion:  2,
		Format:         appendNewLine(newrelicotlpDefaultFormat),
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
					testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(&service, []*gofastly.NewRelicOTLP{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelicotlp.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLNewRelicOTLPConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(&service, []*gofastly.NewRelicOTLP{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_newrelicotlp.#", "2"),
				),
				PreventDiskCleanup: true,
			},
		},
	})
}

func testAccCheckFastlyServiceVCLNewRelicOTLPAttributes(service *gofastly.ServiceDetail, newrelicotlp []*gofastly.NewRelicOTLP, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		newrelicotlpList, err := conn.ListNewRelicOTLP(&gofastly.ListNewRelicOTLPInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up NewRelic OTLP Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(newrelicotlpList) != len(newrelicotlp) {
			return fmt.Errorf("newRelic List count mismatch, expected (%d), got (%d)", len(newrelicotlp), len(newrelicotlpList))
		}

		log.Printf("[DEBUG] newrelicotlpList = %#v\n", newrelicotlpList)

		var found int
		for _, d := range newrelicotlp {
			for _, dl := range newrelicotlpList {
				if d.Name == dl.Name {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ID
					d.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					dl.CreatedAt = nil
					dl.UpdatedAt = nil

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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLNewRelicOTLPConfigUpdate(name, domain string) string {
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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    region = "EU"
}

  logging_newrelicotlp {
    name   = "another-newrelicotlp-endpoint"
    token  = "another-token"
	url    = "https://example.nr-data.net"
	format = <<EOF
`+escapePercentSign(newrelicotlpDefaultFormat)+`
EOF
  }

  force_destroy = true
}
`, name, domain)
}
