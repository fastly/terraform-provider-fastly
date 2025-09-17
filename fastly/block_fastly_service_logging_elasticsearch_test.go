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

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestAccFastlyServiceVCL_logging_elasticsearch_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Elasticsearch{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("elasticsearch-endpoint"),
		Index:             gofastly.ToPointer("#{%F}"),
		URL:               gofastly.ToPointer("https://es.example.com"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		FormatVersion:     gofastly.ToPointer(2),
		Format:            gofastly.ToPointer(LoggingElasticsearchDefaultFormat),
		User:              gofastly.ToPointer("user"),
		Password:          gofastly.ToPointer("password"),
		Pipeline:          gofastly.ToPointer("my-pipeline"),
		TLSCACert:         gofastly.ToPointer(caCert(t)),
		TLSClientCert:     gofastly.ToPointer(certificate(t)),
		TLSClientKey:      gofastly.ToPointer(privateKey(t)),
		TLSHostname:       gofastly.ToPointer("example.com"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Placement:         gofastly.ToPointer("none"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log1AfterUpdate := gofastly.Elasticsearch{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("elasticsearch-endpoint"),
		Index:             gofastly.ToPointer("#{%F}"),
		URL:               gofastly.ToPointer("https://es.example.com"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		FormatVersion:     gofastly.ToPointer(2),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		User:              gofastly.ToPointer("newuser"),
		Password:          gofastly.ToPointer("newpassword"),
		Pipeline:          gofastly.ToPointer("my-new-pipeline"),
		TLSCACert:         gofastly.ToPointer(caCert(t)),
		TLSClientCert:     gofastly.ToPointer(certificate(t)),
		TLSClientKey:      gofastly.ToPointer(privateKey(t)),
		TLSHostname:       gofastly.ToPointer("example.com"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Placement:         gofastly.ToPointer("none"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	log2 := gofastly.Elasticsearch{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("another-elasticsearch-endpoint"),
		Index:             gofastly.ToPointer("#{%F}"),
		URL:               gofastly.ToPointer("https://es2.example.com"),
		RequestMaxBytes:   gofastly.ToPointer(1000),
		RequestMaxEntries: gofastly.ToPointer(0),
		FormatVersion:     gofastly.ToPointer(2),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		User:              gofastly.ToPointer("username"),
		Password:          gofastly.ToPointer("secret-password"),
		Pipeline:          gofastly.ToPointer("my-new-pipeline"),
		TLSCACert:         gofastly.ToPointer(caCert(t)),
		TLSClientCert:     gofastly.ToPointer(certificate(t)),
		TLSClientKey:      gofastly.ToPointer(privateKey(t)),
		TLSHostname:       gofastly.ToPointer("example.com"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Placement:         gofastly.ToPointer("none"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLElasticsearchConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLElasticsearchAttributes(&service, []*gofastly.Elasticsearch{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_elasticsearch.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLElasticsearchConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLElasticsearchAttributes(&service, []*gofastly.Elasticsearch{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_elasticsearch.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_elasticsearch_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Elasticsearch{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("elasticsearch-endpoint"),
		Index:             gofastly.ToPointer("#{%F}"),
		URL:               gofastly.ToPointer("https://es.example.com"),
		RequestMaxBytes:   gofastly.ToPointer(0),
		RequestMaxEntries: gofastly.ToPointer(0),
		User:              gofastly.ToPointer("user"),
		Password:          gofastly.ToPointer("password"),
		Pipeline:          gofastly.ToPointer("my-pipeline"),
		TLSCACert:         gofastly.ToPointer(caCert(t)),
		TLSClientCert:     gofastly.ToPointer(certificate(t)),
		TLSClientKey:      gofastly.ToPointer(privateKey(t)),
		TLSHostname:       gofastly.ToPointer("example.com"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLElasticsearchComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLElasticsearchAttributes(&service, []*gofastly.Elasticsearch{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_elasticsearch.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLElasticsearchAttributes(service *gofastly.ServiceDetail, elasticsearch []*gofastly.Elasticsearch, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		elasticsearchList, err := conn.ListElasticsearch(context.TODO(), &gofastly.ListElasticsearchInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Elasticsearch Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(elasticsearchList) != len(elasticsearch) {
			return fmt.Errorf("elasticsearch List count mismatch, expected (%d), got (%d)", len(elasticsearch), len(elasticsearchList))
		}

		log.Printf("[DEBUG] elasticsearchList = %#v\n", elasticsearchList)

		var found int
		for _, e := range elasticsearch {
			for _, el := range elasticsearchList {
				if gofastly.ToValue(e.Name) == gofastly.ToValue(el.Name) {
					// We don't know these things ahead of time, so populate them now.
					e.ServiceID = service.ServiceID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time.
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
						return fmt.Errorf("bad match Elasticsearch logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(elasticsearch) {
			return fmt.Errorf("error matching Elasticsearch Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLElasticsearchComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-elasticsearch-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_elasticsearch {
    name     = "elasticsearch-endpoint"
    index    = "#{%%F}"
    url      = "https://es.example.com"
    pipeline = "my-pipeline"
    user     = "user"
    password = "password"
    tls_ca_cert       = file("test_fixtures/fastly_test_cacert")
    tls_client_cert   = file("test_fixtures/fastly_test_certificate")
    tls_client_key    = file("test_fixtures/fastly_test_privatekey")
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

func testAccServiceVCLElasticsearchConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-elasticsearch-logging"
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

  logging_elasticsearch {
    name     = "elasticsearch-endpoint"
    index    = "#{%%F}"
    url      = "https://es.example.com"
		pipeline = "my-pipeline"
		user     = "user"
		password = "password"
		tls_ca_cert       = file("test_fixtures/fastly_test_cacert")
		tls_client_cert   = file("test_fixtures/fastly_test_certificate")
		tls_client_key    = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		placement          = "none"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLElasticsearchConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-elasticsearch-logging"
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

  logging_elasticsearch {
    name   = "elasticsearch-endpoint"
    index  = "#{%%F}"
    url    = "https://es.example.com"
    format = %q
		pipeline = "my-new-pipeline"
		user     = "newuser"
		password = "newpassword"
		tls_ca_cert       = file("test_fixtures/fastly_test_cacert")
		tls_client_cert   = file("test_fixtures/fastly_test_certificate")
		tls_client_key    = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		placement          = "none"
  }

  logging_elasticsearch {
    name              = "another-elasticsearch-endpoint"
    index             = "#{%%F}"
    url               = "https://es2.example.com"
    user              = "username"
    password          = "secret-password"
    format            = %q
    request_max_bytes = 1000
		pipeline          = "my-new-pipeline"
		tls_ca_cert       = file("test_fixtures/fastly_test_cacert")
		tls_client_cert   = file("test_fixtures/fastly_test_certificate")
		tls_client_key    = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		placement          = "none"
  }

  force_destroy = true
}
`, name, domain, format, format)
}

func TestResourceFastlyFlattenElasticsearch(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Elasticsearch
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Elasticsearch{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("elasticsearch-endpoint"),
					Index:             gofastly.ToPointer("index"),
					URL:               gofastly.ToPointer("https://logs.example.com"),
					Pipeline:          gofastly.ToPointer("my-pipeline-id"),
					RequestMaxEntries: gofastly.ToPointer(10),
					RequestMaxBytes:   gofastly.ToPointer(10),
					User:              gofastly.ToPointer("user"),
					Password:          gofastly.ToPointer("password"),
					TLSCACert:         gofastly.ToPointer(caCert(t)),
					TLSClientCert:     gofastly.ToPointer(certificate(t)),
					TLSClientKey:      gofastly.ToPointer(privateKey(t)),
					TLSHostname:       gofastly.ToPointer("example.com"),
					ResponseCondition: gofastly.ToPointer("response_condition"),
					Format:            gofastly.ToPointer(LoggingElasticsearchDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
					Placement:         gofastly.ToPointer("none"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":                "elasticsearch-endpoint",
					"index":               "index",
					"url":                 "https://logs.example.com",
					"pipeline":            "my-pipeline-id",
					"user":                "user",
					"password":            "password",
					"tls_ca_cert":         caCert(t),
					"tls_client_cert":     certificate(t),
					"tls_client_key":      privateKey(t),
					"tls_hostname":        "example.com",
					"response_condition":  "response_condition",
					"format":              LoggingElasticsearchDefaultFormat,
					"placement":           "none",
					"request_max_entries": 10,
					"request_max_bytes":   10,
					"format_version":      2,
					"processing_region":   "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenElasticsearch(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
