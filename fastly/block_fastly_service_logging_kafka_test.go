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

func TestResourceFastlyFlattenKafka(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kafka
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Kafka{
				{
					ServiceVersion:    new(1),
					Name:              new("kafka-endpoint"),
					Topic:             new("topic"),
					Brokers:           new("127.0.0.1,127.0.0.2"),
					CompressionCodec:  new("snappy"),
					RequiredACKs:      new("-1"),
					UseTLS:            new(true),
					TLSCACert:         new(caCert(t)),
					TLSClientCert:     new(certificate(t)),
					TLSClientKey:      new(privateKey(t)),
					TLSHostname:       new("example.com"),
					ResponseCondition: new("response_condition"),
					Format:            new(LoggingKafkaDefaultFormat),
					FormatVersion:     new(2),
					Placement:         new("none"),
					ParseLogKeyvals:   new(true),
					RequestMaxBytes:   new(12345),
					AuthMethod:        new("scram-sha-512"),
					User:              new("user"),
					Password:          new("password"),
					ProcessingRegion:  new("eu"),
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
					"format":             LoggingKafkaDefaultFormat,
					"placement":          "none",
					"format_version":     2,
					"parse_log_keyvals":  true,
					"request_max_bytes":  12345,
					"auth_method":        "scram-sha-512",
					"user":               "user",
					"password":           "password",
					"processing_region":  "eu",
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

func TestAccFastlyServiceLoggingKafka_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kafka{
		ServiceVersion:    new(1),
		Name:              new("kafkalogger"),
		Topic:             new("topic"),
		Brokers:           new("127.0.0.1,127.0.0.2"),
		CompressionCodec:  new("snappy"),
		RequiredACKs:      new("-1"),
		UseTLS:            new(true),
		TLSCACert:         new(caCert(t)),
		TLSClientCert:     new(certificate(t)),
		TLSClientKey:      new(privateKey(t)),
		TLSHostname:       new("example.com"),
		ResponseCondition: new("response_condition_test"),
		Format:            new(LoggingKafkaDefaultFormat),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ParseLogKeyvals:   new(true),
		RequestMaxBytes:   new(12345),
		AuthMethod:        new("plain"),
		User:              new("user"),
		Password:          new("password"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Kafka{
		ServiceVersion:    new(1),
		Name:              new("kafkalogger"),
		Topic:             new("newtopic"),
		Brokers:           new("127.0.0.3,127.0.0.4"),
		CompressionCodec:  new("lz4"),
		RequiredACKs:      new("0"),
		UseTLS:            new(false),
		TLSCACert:         new(caCert(t)),
		TLSClientCert:     new(certificate(t)),
		TLSClientKey:      new(privateKey(t)),
		TLSHostname:       new("example2.com"),
		ResponseCondition: new("response_condition_test"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ParseLogKeyvals:   new(true),
		RequestMaxBytes:   new(12345),
		AuthMethod:        new("scram-sha-256"),
		User:              new("user"),
		Password:          new("password"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Kafka{
		ServiceVersion:    new(1),
		Name:              new("kafkalogger2"),
		Topic:             new("topicb"),
		Brokers:           new("127.0.0.3,127.0.0.4"),
		CompressionCodec:  new("gzip"),
		RequiredACKs:      new("1"),
		UseTLS:            new(true),
		TLSCACert:         new(caCert(t)),
		TLSClientCert:     new(certificate(t)),
		TLSClientKey:      new(privateKey(t)),
		TLSHostname:       new("example.com"),
		ResponseCondition: new("response_condition_test"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		Placement:         new("none"),
		ParseLogKeyvals:   new(true),
		RequestMaxBytes:   new(12345),
		AuthMethod:        new("scram-sha-256"),
		User:              new("user"),
		Password:          new("password"),
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

func TestAccFastlyServiceLoggingKafka_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kafka{
		AuthMethod:       new(""),
		Brokers:          new("127.0.0.1,127.0.0.2"),
		CompressionCodec: new("snappy"),
		Name:             new("kafkalogger"),
		ParseLogKeyvals:  new(false),
		Password:         new(""),
		RequestMaxBytes:  new(0),
		RequiredACKs:     new("-1"),
		ServiceVersion:   new(1),
		TLSCACert:        new(caCert(t)),
		TLSClientCert:    new(certificate(t)),
		TLSClientKey:     new(privateKey(t)),
		TLSHostname:      new("example.com"),
		Topic:            new("topic"),
		UseTLS:           new(true),
		User:             new(""),
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
				Config: testAccServiceVCLKafkaComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLKafkaAttributes(&service, []*gofastly.Kafka{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_kafka.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingKafka_BooleanFieldsPreservedOnUpdate(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := acctest.RandomWithPrefix("tf-kafka")
	domainName := fmt.Sprintf("test.%s.com", acctest.RandString(10))

	kafkaName := "kafka-logging-preserve"

	useTLS := true
	parseLogKeyvals := true

	initialMaxBytes := 10000
	updatedMaxBytes := 20000

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLKafkaLoggingPreserveBooleansWithMaxBytes(
					serviceName, domainName, kafkaName, useTLS, parseLogKeyvals, initialMaxBytes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.use_tls", fmt.Sprintf("%t", useTLS)),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.parse_log_keyvals", fmt.Sprintf("%t", parseLogKeyvals)),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.request_max_bytes", fmt.Sprintf("%d", initialMaxBytes)),
				),
			},
			{
				// Only update non-ForceNew field: request_max_bytes
				Config: testAccServiceVCLKafkaLoggingPreserveBooleansWithMaxBytes(
					serviceName, domainName, kafkaName, useTLS, parseLogKeyvals, updatedMaxBytes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.use_tls", fmt.Sprintf("%t", useTLS)),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.parse_log_keyvals", fmt.Sprintf("%t", parseLogKeyvals)),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kafka.0.request_max_bytes", fmt.Sprintf("%d", updatedMaxBytes)),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLKafkaAttributes(service *gofastly.ServiceDetail, kafka []*gofastly.Kafka, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		kafkaList, err := conn.ListKafkas(context.TODO(), &gofastly.ListKafkasInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Kafka Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(kafkaList) != len(kafka) {
			return fmt.Errorf("kafka List count mismatch, expected (%d), got (%d)", len(kafka), len(kafkaList))
		}

		log.Printf("[DEBUG] kafkaList = %#v\n", kafkaList)

		var found int
		for _, s := range kafka {
			for _, sl := range kafkaList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(sl.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ServiceID
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
	  	topic  			   = "topic"
		brokers            = "127.0.0.1,127.0.0.2"
		compression_codec  = "snappy"
		required_acks      = "-1"
		use_tls            = true
		tls_ca_cert        = file("test_fixtures/fastly_test_cacert")
		tls_client_cert    = file("test_fixtures/fastly_test_certificate")
		tls_client_key     = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
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
		format_version     = 2
		placement          = "none"
		parse_log_keyvals  = true
		request_max_bytes  = 12345
		auth_method        = "plain"
		user               = "user"
		password           = "password"
    processing_region = "us"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKafkaConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
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
		format             = %q
		format_version     = 2
		placement          = "none"
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
		format             = %q
		format_version     = 2
		placement          = "none"
		parse_log_keyvals  = true
		request_max_bytes  = 12345
		auth_method        = "scram-sha-256"
		user               = "user"
		password           = "password"
	}

	force_destroy = true
}`, name, domain, format, format)
}

func testAccServiceVCLKafkaLoggingPreserveBooleansWithMaxBytes(
	serviceName, domainName, kafkaName string,
	useTLS, parseLogKeyvals bool,
	requestMaxBytes int,
) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "test"
  }

  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }

  logging_kafka {
    name              = "%s"
    topic             = "test-topic"
    brokers           = "127.0.0.1:9092"
    required_acks     = "1"
    use_tls           = %t
    parse_log_keyvals = %t
    request_max_bytes = %d
  }

  force_destroy = true
}
`, serviceName, domainName, kafkaName, useTLS, parseLogKeyvals, requestMaxBytes)
}
