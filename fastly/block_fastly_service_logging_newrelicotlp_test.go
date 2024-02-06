package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenNewRelicOTLP(t *testing.T) {
	var format, placement, responseCondition, loggingURL *string
	cases := []struct {
		remote []*gofastly.NewRelicOTLP
		local  []map[string]any
	}{
		{
			remote: []*gofastly.NewRelicOTLP{
				{
					ServiceVersion: gofastly.ToPointer(1),
					Name:           gofastly.ToPointer("newrelicotlp-endpoint"),
					Token:          gofastly.ToPointer("token"),
					Region:         gofastly.ToPointer("US"),
					FormatVersion:  gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"format":             format, // implies nil
					"format_version":     gofastly.ToPointer(2),
					"name":               gofastly.ToPointer("newrelicotlp-endpoint"),
					"placement":          placement, // implies nil
					"region":             gofastly.ToPointer("US"),
					"response_condition": responseCondition, // implies nil
					"token":              gofastly.ToPointer("token"),
					"url":                loggingURL, // implies nil
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
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("token"),
		Region:         gofastly.ToPointer("US"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
		URL:               gofastly.ToPointer(""),
	}

	log1AfterUpdate := gofastly.NewRelicOTLP{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("t0k3n"),
		Region:         gofastly.ToPointer("EU"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
		URL:               gofastly.ToPointer(""),
	}

	log2 := gofastly.NewRelicOTLP{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("another-newrelicotlp-endpoint"),
		Token:          gofastly.ToPointer("another-token"),
		Region:         gofastly.ToPointer("US"),
		URL:            gofastly.ToPointer("https://example.nr-data.net"),
		FormatVersion:  gofastly.ToPointer(2),
		Format:         gofastly.ToPointer(appendNewLine(newrelicotlpDefaultFormat)),
		// The Fastly API returns an empty string if nothing set by the user (it should probably set null)
		ResponseCondition: gofastly.ToPointer(""),
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
