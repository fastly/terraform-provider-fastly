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

func TestAccFastlyServiceLoggingGrafanaCloudLogs_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.GrafanaCloudLogs{
		Format:            new(LoggingGrafanaCloudLogsDefaultFormat),
		FormatVersion:     new(2),
		Name:              new("grafanacloudlogs-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		User:              new("123456"),
		Token:             new("token"),
		URL:               new("https://test123.grafana.net"),
		Index:             new("{\"label\": \"value\"}"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.GrafanaCloudLogs{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("grafanacloudlogs-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		User:              new("987654"),
		Token:             new("t0k3n"),
		URL:               new("https://test456.grafana.net"),
		Index:             new("{\"label2\": \"value2\"}"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.GrafanaCloudLogs{
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Name:              new("another-grafanacloudlogs-endpoint"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		User:              new("123456"),
		URL:               new("https://test789.grafana.net"),
		Index:             new("{\"label3\": \"value3\"}"),
		Token:             new("another-token"),
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

func TestAccFastlyServiceLoggingGrafanaCloudLogs_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.GrafanaCloudLogs{
		ServiceVersion:   new(1),
		Name:             new("grafanacloudlogs-endpoint"),
		User:             new("123456"),
		Token:            new("token"),
		URL:              new("https://test123.grafana.net"),
		Index:            new("{\"label\": \"value\"}"),
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
		grafanacloudlogsList, err := conn.ListGrafanaCloudLogs(context.TODO(), &gofastly.ListGrafanaCloudLogsInput{
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
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLGrafanaCloudLogsConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
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
    format = %q
  }

  logging_grafanacloudlogs {
    name  = "another-grafanacloudlogs-endpoint"
    token = "another-token"
	user   = "123456"
	url    = "https://test789.grafana.net"
	index  = "{\"label3\": \"value3\"}"
		format = %q
  }

  force_destroy = true
}
`, name, domain, format, format)
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

func TestResourceFastlyFlattenGrafanaCloudLogs(t *testing.T) {
	cases := []struct {
		remote []*gofastly.GrafanaCloudLogs
		local  []map[string]any
	}{
		{
			remote: []*gofastly.GrafanaCloudLogs{
				{
					ServiceVersion:   new(1),
					Name:             new("grafanacloudlogs-endpoint"),
					User:             new("123456"),
					Token:            new("token"),
					URL:              new("https://test123.grafana.net"),
					Index:            new("{\"label\": \"value\"}"),
					Format:           new(LoggingGrafanaCloudLogsDefaultFormat),
					FormatVersion:    new(2),
					ProcessingRegion: new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":              "grafanacloudlogs-endpoint",
					"user":              "123456",
					"token":             "token",
					"url":               "https://test123.grafana.net",
					"index":             "{\"label\": \"value\"}",
					"format":            LoggingGrafanaCloudLogsDefaultFormat,
					"format_version":    2,
					"processing_region": "eu",
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
