package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenKafka(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kafka
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Kafka{
				{
					Version:           1,
					Name:              "kafka-endpoint",
					Topic:             "topic",
					Brokers:           "127.0.0.1,127.0.0.2",
					CompressionCodec:  "snappy",
					RequiredACKs:      "-1",
					UseTLS:            true,
					TLSCACert:         caCert(t),
					TLSClientCert:     certificate(t),
					TLSClientKey:      privateKey(t),
					TLSHostname:       "example.com",
					ResponseCondition: "response_condition",
					Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
					FormatVersion:     2,
					Placement:         "none",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "kafka-endpoint",
					"topic":              "topic",
					"brokers":            "127.0.0.1,127.0.0.2",
					"compression_codec":  "snappy",
					"required_acks":      "-1",
					"use_tls":            true,
					"tls_ca_cert":        caCert(t),
					"tls_client_cert":    certificate(t),
					"tls_client_key":     privateKey(t),
					"tls_hostname":       "example.com",
					"response_condition": "response_condition",
					"format":             `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":          "none",
					"format_version":     uint(2),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenKafka(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceV1_kafkalogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	nameCompute := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1Compute := gofastly.Kafka{
		Version:          1,
		Name:             "kafkalogger",
		Topic:            "topic",
		Brokers:          "127.0.0.1,127.0.0.2",
		CompressionCodec: "snappy",
		RequiredACKs:     "-1",
		UseTLS:           true,
		TLSCACert:        caCert(t),
		TLSClientCert:    certificate(t),
		TLSClientKey:     privateKey(t),
		TLSHostname:      "example.com",
	}

	log1 := gofastly.Kafka{
		Version:           1,
		Name:              "kafkalogger",
		Topic:             "topic",
		Brokers:           "127.0.0.1,127.0.0.2",
		CompressionCodec:  "snappy",
		RequiredACKs:      "-1",
		UseTLS:            true,
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "none",
	}

	log1_after_update := gofastly.Kafka{
		Version:           1,
		Name:              "kafkalogger",
		Topic:             "newtopic",
		Brokers:           "127.0.0.3,127.0.0.4",
		CompressionCodec:  "lz4",
		RequiredACKs:      "0",
		UseTLS:            false,
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example2.com",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "waf_debug",
	}

	log2 := gofastly.Kafka{
		Version:           1,
		Name:              "kafkalogger2",
		Topic:             "topicb",
		Brokers:           "127.0.0.3,127.0.0.4",
		CompressionCodec:  "gzip",
		RequiredACKs:      "1",
		UseTLS:            true,
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
		ResponseCondition: "response_condition_test",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		FormatVersion:     2,
		Placement:         "none",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1KafkaComputeConfig(nameCompute, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1KafkaAttributes(&service, []*gofastly.Kafka{&log1Compute}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", nameCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_kafka.#", "1"),
				),
			},

			{
				Config: testAccServiceV1KafkaConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1KafkaAttributes(&service, []*gofastly.Kafka{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_kafka.#", "1"),
				),
			},

			{
				Config: testAccServiceV1KafkaConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1KafkaAttributes(&service, []*gofastly.Kafka{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_kafka.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1KafkaAttributes(service *gofastly.ServiceDetail, kafka []*gofastly.Kafka, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		kafkaList, err := conn.ListKafkas(&gofastly.ListKafkasInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Kafka Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(kafkaList) != len(kafka) {
			return fmt.Errorf("Kafka List count mismatch, expected (%d), got (%d)", len(kafka), len(kafkaList))
		}

		log.Printf("[DEBUG] kafkaList = %#v\n", kafkaList)

		var found int
		for _, s := range kafka {
			for _, sl := range kafkaList {
				if s.Name == sl.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					sl.CreatedAt = nil
					sl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						sl.FormatVersion = s.FormatVersion
						sl.Format = s.Format
						sl.ResponseCondition = s.ResponseCondition
						sl.Placement = s.Placement
					}

					if diff := cmp.Diff(s, sl); diff != "" {
						return fmt.Errorf("Bad match Kafka logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(kafka) {
			return fmt.Errorf("Error matching Kafka Logging rules")
		}

		return nil
	}
}

func testAccServiceV1KafkaComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-kafka-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_kafka {
		name               = "kafkalogger"
	  topic  						 = "topic"
		brokers            = "127.0.0.1,127.0.0.2"
		compression_codec  = "snappy"
		required_acks      = "-1"
		use_tls            = true
		tls_ca_cert        = file("test_fixtures/fastly_test_cacert")
		tls_client_cert    = file("test_fixtures/fastly_test_certificate")
		tls_client_key     = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
	}

	package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1KafkaConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-kafka-logging"
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

	logging_kafka {
		name               = "kafkalogger"
	  topic  						 = "topic"
		brokers            = "127.0.0.1,127.0.0.2"
		compression_codec  = "snappy"
		required_acks      = "-1"
		use_tls            = true
		tls_ca_cert        = file("test_fixtures/fastly_test_cacert")
		tls_client_cert    = file("test_fixtures/fastly_test_certificate")
		tls_client_key     = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "none"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1KafkaConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-testing-domain"
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

	logging_kafka {
		name               = "kafkalogger"
	  topic  						 = "newtopic"
		brokers            = "127.0.0.3,127.0.0.4"
		compression_codec  = "lz4"
		required_acks      = "0"
		use_tls            = false
		tls_ca_cert        = file("test_fixtures/fastly_test_cacert")
		tls_client_cert    = file("test_fixtures/fastly_test_certificate")
		tls_client_key     = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example2.com"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "waf_debug"
	}

	logging_kafka {
		name               = "kafkalogger2"
	  	topic  			   = "topicb"
		brokers            = "127.0.0.3,127.0.0.4"
		compression_codec  = "gzip"
		required_acks      = "1"
		use_tls            = true
		tls_ca_cert        = file("test_fixtures/fastly_test_cacert")
		tls_client_cert    = file("test_fixtures/fastly_test_certificate")
		tls_client_key     = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version     = 2
		placement          = "none"
	}

	force_destroy = true
}`, name, domain)
}
