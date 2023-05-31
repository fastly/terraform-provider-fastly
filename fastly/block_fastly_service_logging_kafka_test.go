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

func TestResourceFastlyFlattenKafka(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kafka
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Kafka{
				{
					ServiceVersion:    1,
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
					ParseLogKeyvals:   true,
					RequestMaxBytes:   12345,
					AuthMethod:        "scram-sha-512",
					User:              "user",
					Password:          "password",
				},
			},
			local: []map[string]any{
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
					"format_version":     2,
					"parse_log_keyvals":  true,
					"request_max_bytes":  12345,
					"auth_method":        "scram-sha-512",
					"user":               "user",
					"password":           "password",
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

func TestAccFastlyServiceVCL_kafkalogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kafka{
		ServiceVersion:    1,
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
		ParseLogKeyvals:   true,
		RequestMaxBytes:   12345,
		AuthMethod:        "plain",
		User:              "user",
		Password:          "password",
	}

	log1AfterUpdate := gofastly.Kafka{
		ServiceVersion:    1,
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
		ParseLogKeyvals:   true,
		RequestMaxBytes:   12345,
		AuthMethod:        "scram-sha-256",
		User:              "user",
		Password:          "password",
	}

	log2 := gofastly.Kafka{
		ServiceVersion:    1,
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
		ParseLogKeyvals:   true,
		RequestMaxBytes:   12345,
		AuthMethod:        "scram-sha-256",
		User:              "user",
		Password:          "password",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLKafkaConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKafkaAttributes(&service, []*gofastly.Kafka{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_kafka.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLKafkaConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKafkaAttributes(&service, []*gofastly.Kafka{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_kafka.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_kafkalogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kafka{
		ServiceVersion:   1,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLKafkaComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLKafkaAttributes(&service, []*gofastly.Kafka{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_kafka.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLKafkaAttributes(service *gofastly.ServiceDetail, kafka []*gofastly.Kafka, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		kafkaList, err := conn.ListKafkas(&gofastly.ListKafkasInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Kafka Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(kafkaList) != len(kafka) {
			return fmt.Errorf("kafka List count mismatch, expected (%d), got (%d)", len(kafka), len(kafkaList))
		}

		log.Printf("[DEBUG] kafkaList = %#v\n", kafkaList)

		var found int
		for _, s := range kafka {
			for _, sl := range kafkaList {
				if s.Name == sl.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
						return fmt.Errorf("bad match Kafka logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(kafka) {
			return fmt.Errorf("error matching Kafka Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLKafkaComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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
    source_code_hash = data.fastly_package_hash.example.hash
  }

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKafkaConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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
		parse_log_keyvals  = true
		request_max_bytes  = 12345
		auth_method        = "plain"
		user               = "user"
		password           = "password"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKafkaConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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
		parse_log_keyvals  = true
		request_max_bytes  = 12345
		auth_method        = "scram-sha-256"
		user               = "user"
		password           = "password"
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
		parse_log_keyvals  = true
		request_max_bytes  = 12345
		auth_method        = "scram-sha-256"
		user               = "user"
		password           = "password"
	}

	force_destroy = true
}`, name, domain)
}
