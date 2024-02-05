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

func TestResourceFastlyFlattenDatadog(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Datadog
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Datadog{
				{
					ServiceVersion: gofastly.ToPointer(1),
					Name:           gofastly.ToPointer("datadog-endpoint"),
					Token:          gofastly.ToPointer("token"),
					Region:         gofastly.ToPointer("US"),
					FormatVersion:  gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":           "datadog-endpoint",
					"token":          "token",
					"region":         "US",
					"format_version": 2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDatadog(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

var datadogDefaultFormat = `{
  "ddsource": "fastly",
  "service": "%{req.service_id}V",
  "date": "%{begin:%Y-%m-%dT%H:%M:%S%Z}t",
  "time_start": "%{begin:%Y-%m-%dT%H:%M:%S%Z}t",
  "time_end": "%{end:%Y-%m-%dT%H:%M:%S%Z}t",
  "http": {
    "request_time_ms": %D,
    "method": "%m",
    "url": "%{json.escape(req.url)}V",
    "useragent": "%{User-Agent}i",
    "referer": "%{Referer}i",
    "protocol": "%H",
    "request_x_forwarded_for": "%{X-Forwarded-For}i",
    "status_code": "%s"
  },
  "network": {
    "client": {
     "ip": "%h",
     "name": "%{client.as.name}V",
     "number": "%{client.as.number}V",
     "connection_speed": "%{client.geo.conn_speed}V"
    },
   "destination": {
     "ip": "%A"
    },
  "geoip": {
  "geo_city": "%{client.geo.city.utf8}V",
  "geo_country_code": "%{client.geo.country_code}V",
  "geo_continent_code": "%{client.geo.continent_code}V",
  "geo_region": "%{client.geo.region}V"
  },
  "bytes_written": %B,
  "bytes_read": %{req.body_bytes_read}V
  },
  "host": "%{Fastly-Orig-Host}i",
  "origin_host": "%v",
  "is_ipv6": %{if(req.is_ipv6, "true", "false")}V,
  "is_tls": %{if(req.is_ssl, "true", "false")}V,
  "tls_client_protocol": "%{json.escape(tls.client.protocol)}V",
  "tls_client_servername": "%{json.escape(tls.client.servername)}V",
  "tls_client_cipher": "%{json.escape(tls.client.cipher)}V",
  "tls_client_cipher_sha": "%{json.escape(tls.client.ciphers_sha)}V",
  "tls_client_tlsexts_sha": "%{json.escape(tls.client.tlsexts_sha)}V",
  "is_h2": %{if(fastly_info.is_h2, "true", "false")}V,
  "is_h2_push": %{if(fastly_info.h2.is_push, "true", "false")}V,
  "h2_stream_id": "%{fastly_info.h2.stream_id}V",
  "request_accept_content": "%{Accept}i",
  "request_accept_language": "%{Accept-Language}i",
  "request_accept_encoding": "%{Accept-Encoding}i",
  "request_accept_charset": "%{Accept-Charset}i",
  "request_connection": "%{Connection}i",
  "request_dnt": "%{DNT}i",
  "request_forwarded": "%{Forwarded}i",
  "request_via": "%{Via}i",
  "request_cache_control": "%{Cache-Control}i",
  "request_x_requested_with": "%{X-Requested-With}i",
  "request_x_att_device_id": "%{X-ATT-Device-Id}i",
  "content_type": "%{Content-Type}o",
  "is_cacheable": %{if(fastly_info.state~"^(HIT|MISS)$", "true","false")}V,
  "response_age": "%{Age}o",
  "response_cache_control": "%{Cache-Control}o",
  "response_expires": "%{Expires}o",
  "response_last_modified": "%{Last-Modified}o",
  "response_tsv": "%{TSV}o",
  "server_datacenter": "%{server.datacenter}V",
  "req_header_size": %{req.header_bytes_read}V,
  "resp_header_size": %{resp.header_bytes_written}V,
  "socket_cwnd": %{client.socket.cwnd}V,
  "socket_nexthop": "%{client.socket.nexthop}V",
  "socket_tcpi_rcv_mss": %{client.socket.tcpi_rcv_mss}V,
  "socket_tcpi_snd_mss": %{client.socket.tcpi_snd_mss}V,
  "socket_tcpi_rtt": %{client.socket.tcpi_rtt}V,
  "socket_tcpi_rttvar": %{client.socket.tcpi_rttvar}V,
  "socket_tcpi_rcv_rtt": %{client.socket.tcpi_rcv_rtt}V,
  "socket_tcpi_rcv_space": %{client.socket.tcpi_rcv_space}V,
  "socket_tcpi_last_data_sent": %{client.socket.tcpi_last_data_sent}V,
  "socket_tcpi_total_retrans": %{client.socket.tcpi_total_retrans}V,
  "socket_tcpi_delta_retrans": %{client.socket.tcpi_delta_retrans}V,
  "socket_ploss": %{client.socket.ploss}V
}`

func TestAccFastlyServiceVCL_logging_datadog_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("datadog-endpoint"),
		Region:            gofastly.ToPointer("US"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("token"),
	}

	log1AfterUpdate := gofastly.Datadog{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("datadog-endpoint"),
		Region:            gofastly.ToPointer("EU"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("t0k3n"),
	}

	log2 := gofastly.Datadog{
		Format:            gofastly.ToPointer(datadogDefaultFormat + "\n"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-datadog-endpoint"),
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
				Config: testAccServiceVCLDatadogConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_datadog.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLDatadogConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_datadog.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_datadog_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("datadog-endpoint"),
		Token:          gofastly.ToPointer("token"),
		Region:         gofastly.ToPointer("US"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLDatadogComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLDatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_datadog.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLDatadogAttributes(service *gofastly.ServiceDetail, datadog []*gofastly.Datadog, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		datadogList, err := conn.ListDatadog(&gofastly.ListDatadogInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Datadog Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(datadogList) != len(datadog) {
			return fmt.Errorf("datadog List count mismatch, expected (%d), got (%d)", len(datadog), len(datadogList))
		}

		log.Printf("[DEBUG] datadogList = %#v\n", datadogList)

		var found int
		for _, d := range datadog {
			for _, dl := range datadogList {
				if gofastly.ToValue(d.Name) == gofastly.ToValue(dl.Name) {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ID
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
						return fmt.Errorf("bad match Datadog logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(datadog) {
			return fmt.Errorf("error matching Datadog Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLDatadogConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "token"
    region = "US"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLDatadogConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "t0k3n"
    region = "EU"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
  }

  logging_datadog {
    name  = "another-datadog-endpoint"
    token = "another-token"
		format = <<EOF
`+escapePercentSign(datadogDefaultFormat)+`
EOF
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLDatadogComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-datadog-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_datadog {
    name   = "datadog-endpoint"
    token  = "token"
    region = "US"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
