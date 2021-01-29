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

func TestResourceFastlyFlattenDatadog(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Datadog
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Datadog{
				{
					ServiceVersion: 1,
					Name:           "datadog-endpoint",
					Token:          "token",
					Region:         "US",
					FormatVersion:  2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":           "datadog-endpoint",
					"token":          "token",
					"region":         "US",
					"format_version": uint(2),
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

func TestAccFastlyServiceV1_logging_datadog_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		ServiceVersion: 1,
		Name:           "datadog-endpoint",
		Token:          "token",
		Region:         "US",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	log1_after_update := gofastly.Datadog{
		ServiceVersion: 1,
		Name:           "datadog-endpoint",
		Token:          "t0k3n",
		Region:         "EU",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.Datadog{
		ServiceVersion: 1,
		Name:           "another-datadog-endpoint",
		Token:          "another-token",
		Region:         "US",
		FormatVersion:  2,
		Format:         datadogDefaultFormat + "\n",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DatadogConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_datadog.#", "1"),
				),
			},

			{
				Config: testAccServiceV1DatadogConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DatadogAttributes(&service, []*gofastly.Datadog{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_datadog.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_datadog_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Datadog{
		ServiceVersion: 1,
		Name:           "datadog-endpoint",
		Token:          "token",
		Region:         "US",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DatadogComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1DatadogAttributes(&service, []*gofastly.Datadog{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_datadog.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DatadogAttributes(service *gofastly.ServiceDetail, datadog []*gofastly.Datadog, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		datadogList, err := conn.ListDatadog(&gofastly.ListDatadogInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Datadog Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(datadogList) != len(datadog) {
			return fmt.Errorf("Datadog List count mismatch, expected (%d), got (%d)", len(datadog), len(datadogList))
		}

		log.Printf("[DEBUG] datadogList = %#v\n", datadogList)

		var found int
		for _, d := range datadog {
			for _, dl := range datadogList {
				if d.Name == dl.Name {
					// we don't know these things ahead of time, so populate them now
					d.ServiceID = service.ID
					d.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
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
						return fmt.Errorf("Bad match Datadog logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(datadog) {
			return fmt.Errorf("Error matching Datadog Logging rules")
		}

		return nil
	}
}

func testAccServiceV1DatadogConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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

func testAccServiceV1DatadogConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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

func testAccServiceV1DatadogComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}
