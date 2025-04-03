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

func TestResourceFastlyFlattenGrafanaCloudLogs(t *testing.T) {
	cases := []struct {
		remote []*gofastly.GrafanaCloudLogs
		local  []map[string]any
	}{
		{
			remote: []*gofastly.GrafanaCloudLogs{
				{
					ServiceVersion: gofastly.ToPointer(1),
					Name:           gofastly.ToPointer("grafanacloudlogs-endpoint"),
					User:           gofastly.ToPointer("123456"),
					Token:          gofastly.ToPointer("token"),
					URL:            gofastly.ToPointer("https://test123.grafana.net"),
					Index:          gofastly.ToPointer("{\"label\": \"value\"}"),

					FormatVersion: gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":           "grafanacloudlogs-endpoint",
					"user":           "123456",
					"token":          "token",
					"url":            "https://test123.grafana.net",
					"index":          "{\"label\": \"value\"}",
					"format_version": 2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenGrafanaCloudLogs(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

var grafanacloudlogsDefaultFormat = `{
  "timestamp": "%{strftime(\{"%Y-%m-%dT%H:%M:%S"\}, time.start)}V",
  "client_ip": "%{req.http.Fastly-Client-IP}V",
  "geo_country": "%{client.geo.country_name}V",
  "geo_city": "%{client.geo.city}V",
  "host": "%{if(req.http.Fastly-Orig-Host, req.http.Fastly-Orig-Host, req.http.Host)}V",
  "url": "%{json.escape(req.url)}V",
  "request_method": "%{json.escape(req.method)}V",
  "request_protocol": "%{json.escape(req.proto)}V",
  "request_referer": "%{json.escape(req.http.referer)}V",
  "request_user_agent": "%{json.escape(req.http.User-Agent)}V",
  "response_state": "%{json.escape(fastly_info.state)}V",
  "response_status": %{resp.status}V,
  "response_reason": %{if(resp.response, "%22"+json.escape(resp.response)+"%22", "null")}V,
  "response_body_size": %{resp.body_bytes_written}V,
  "fastly_server": "%{json.escape(server.identity)}V",
  "fastly_is_edge": %{if(fastly.ff.visits_this_service == 0, "true", "false")}V
}`

func TestAccFastlyServiceVCL_logging_grafanacloudlogs_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.GrafanaCloudLogs{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("grafanacloudlogs-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		User:              gofastly.ToPointer("123456"),
		Token:             gofastly.ToPointer("token"),
		URL:               gofastly.ToPointer("https://test123.grafana.net"),
		Index:             gofastly.ToPointer("{\"label\": \"value\"}"),
	}

	log1AfterUpdate := gofastly.GrafanaCloudLogs{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("grafanacloudlogs-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		User:              gofastly.ToPointer("987654"),
		Token:             gofastly.ToPointer("t0k3n"),
		URL:               gofastly.ToPointer("https://test456.grafana.net"),
		Index:             gofastly.ToPointer("{\"label2\": \"value2\"}"),
	}

	log2 := gofastly.GrafanaCloudLogs{
		Format:            gofastly.ToPointer(grafanacloudlogsDefaultFormat + "\n"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-grafanacloudlogs-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		User:              gofastly.ToPointer("123456"),
		URL:               gofastly.ToPointer("https://test789.grafana.net"),
		Index:             gofastly.ToPointer("{\"label3\": \"value3\"}"),
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
				Config: testAccServiceVCLGrafanaCloudLogsConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGrafanaCloudLogsAttributes(&service, []*gofastly.GrafanaCloudLogs{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_grafanacloudlogs.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLGrafanaCloudLogsConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGrafanaCloudLogsAttributes(&service, []*gofastly.GrafanaCloudLogs{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_grafanacloudlogs.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_grafanacloudlogs_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.GrafanaCloudLogs{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("grafanacloudlogs-endpoint"),
		User:           gofastly.ToPointer("123456"),
		Token:          gofastly.ToPointer("token"),
		URL:            gofastly.ToPointer("https://test123.grafana.net"),
		Index:          gofastly.ToPointer("{\"label\": \"value\"}"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLGrafanaCloudLogsComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLGrafanaCloudLogsAttributes(&service, []*gofastly.GrafanaCloudLogs{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_grafanacloudlogs.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLGrafanaCloudLogsAttributes(service *gofastly.ServiceDetail, grafanacloudlogs []*gofastly.GrafanaCloudLogs, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		grafanacloudlogsList, err := conn.ListGrafanaCloudLogs(&gofastly.ListGrafanaCloudLogsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up GrafanaCloudLogs Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(grafanacloudlogsList) != len(grafanacloudlogs) {
			return fmt.Errorf("grafanacloudlogs List count mismatch, expected (%d), got (%d)", len(grafanacloudlogs), len(grafanacloudlogsList))
		}

		log.Printf("[DEBUG] grafanacloudlogsList = %#v\n", grafanacloudlogsList)

		var found int
		for _, d := range grafanacloudlogs {
			for _, dl := range grafanacloudlogsList {
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
						return fmt.Errorf("bad match GrafanaCloudLogs logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(grafanacloudlogs) {
			return fmt.Errorf("error matching GrafanaCloudLogs Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLGrafanaCloudLogsConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-grafanacloudlogs-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_grafanacloudlogs {
    name   = "grafanacloudlogs-endpoint"
	user   = "123456"
	token  = "token"
	url    = "https://test123.grafana.net"
	index  = "{\"label\": \"value\"}"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLGrafanaCloudLogsConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-grafanacloudlogs-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_grafanacloudlogs {
    name   = "grafanacloudlogs-endpoint"
	user   = "987654"
	token  = "t0k3n"
	url    = "https://test456.grafana.net"
	index  = "{\"label2\": \"value2\"}"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
  }

  logging_grafanacloudlogs {
    name  = "another-grafanacloudlogs-endpoint"
    token = "another-token"
	user   = "123456"
	url    = "https://test789.grafana.net"
	index  = "{\"label3\": \"value3\"}"
		format = <<EOF
`+escapePercentSign(grafanacloudlogsDefaultFormat)+`
EOF
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLGrafanaCloudLogsComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-grafanacloudlogs-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_grafanacloudlogs {
    name   = "grafanacloudlogs-endpoint"
	user   = "123456"
    token  = "token"
	url    = "https://test123.grafana.net"
	index  = "{\"label\": \"value\"}"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
