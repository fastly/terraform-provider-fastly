package fastly

import (
	"fmt"
	"log"
	"testing"

	fst "github.com/fastly/go-fastly/v6/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenElasticsearch(t *testing.T) {
	cases := []struct {
		remote []*fst.Elasticsearch
		local  []map[string]interface{}
	}{
		{
			remote: []*fst.Elasticsearch{
				{
					ServiceVersion:    1,
					Name:              "elasticsearch-endpoint",
					Index:             "index",
					URL:               "https://logs.example.com",
					Pipeline:          "my-pipeline-id",
					RequestMaxEntries: 10,
					RequestMaxBytes:   10,
					User:              "user",
					Password:          "password",
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
					"format":              `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":           "none",
					"request_max_entries": uint(10),
					"request_max_bytes":   uint(10),
					"format_version":      uint(2),
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

func TestAccFastlyServiceV1_logging_elasticsearch_basic(t *testing.T) {
	var service fst.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := fst.Elasticsearch{
		ServiceVersion:    1,
		Name:              "elasticsearch-endpoint",
		Index:             "#{%F}",
		URL:               "https://es.example.com",
		RequestMaxBytes:   0,
		RequestMaxEntries: 0,
		FormatVersion:     2,
		Format:            "%h %l %u %t \"%r\" %>s %b",
		User:              "user",
		Password:          "password",
		Pipeline:          "my-pipeline",
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
		ResponseCondition: "response_condition_test",
		Placement:         "none",
	}

	log1_after_update := fst.Elasticsearch{
		ServiceVersion:    1,
		Name:              "elasticsearch-endpoint",
		Index:             "#{%F}",
		URL:               "https://es.example.com",
		RequestMaxBytes:   0,
		RequestMaxEntries: 0,
		FormatVersion:     2,
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
		User:              "newuser",
		Password:          "newpassword",
		Pipeline:          "my-new-pipeline",
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
		ResponseCondition: "response_condition_test",
		Placement:         "none",
	}

	log2 := fst.Elasticsearch{
		ServiceVersion:    1,
		Name:              "another-elasticsearch-endpoint",
		Index:             "#{%F}",
		URL:               "https://es2.example.com",
		RequestMaxBytes:   1000,
		RequestMaxEntries: 0,
		FormatVersion:     2,
		Format:            "%h %l %u %t \"%r\" %>s %b",
		User:              "username",
		Password:          "secret-password",
		Pipeline:          "my-new-pipeline",
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
		ResponseCondition: "response_condition_test",
		Placement:         "none",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{

			{
				Config: testAccServiceV1ElasticsearchConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ElasticsearchAttributes(&service, []*fst.Elasticsearch{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_elasticsearch.#", "1"),
				),
			},

			{
				Config: testAccServiceV1ElasticsearchConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ElasticsearchAttributes(&service, []*fst.Elasticsearch{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_elasticsearch.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_elasticsearch_basic_compute(t *testing.T) {
	var service fst.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := fst.Elasticsearch{
		ServiceVersion:    1,
		Name:              "elasticsearch-endpoint",
		Index:             "#{%F}",
		URL:               "https://es.example.com",
		RequestMaxBytes:   0,
		RequestMaxEntries: 0,
		User:              "user",
		Password:          "password",
		Pipeline:          "my-pipeline",
		TLSCACert:         caCert(t),
		TLSClientCert:     certificate(t),
		TLSClientKey:      privateKey(t),
		TLSHostname:       "example.com",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1ElasticsearchComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1ElasticsearchAttributes(&service, []*fst.Elasticsearch{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_elasticsearch.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1ElasticsearchAttributes(service *fst.ServiceDetail, elasticsearch []*fst.Elasticsearch, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		elasticsearchList, err := conn.ListElasticsearch(&fst.ListElasticsearchInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Elasticsearch Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(elasticsearchList) != len(elasticsearch) {
			return fmt.Errorf("Elasticsearch List count mismatch, expected (%d), got (%d)", len(elasticsearch), len(elasticsearchList))
		}

		log.Printf("[DEBUG] elasticsearchList = %#v\n", elasticsearchList)

		var found int
		for _, e := range elasticsearch {
			for _, el := range elasticsearchList {
				if e.Name == el.Name {
					// We don't know these things ahead of time, so populate them now.
					e.ServiceID = service.ID
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
						return fmt.Errorf("Bad match Elasticsearch logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(elasticsearch) {
			return fmt.Errorf("Error matching Elasticsearch Logging rules")
		}

		return nil
	}
}

func testAccServiceV1ElasticsearchComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1ElasticsearchConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    format   = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
		tls_ca_cert       = file("test_fixtures/fastly_test_cacert")
		tls_client_cert   = file("test_fixtures/fastly_test_certificate")
		tls_client_key    = file("test_fixtures/fastly_test_privatekey")
		tls_hostname       = "example.com"
		response_condition = "response_condition_test"
		placement          = "none"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1ElasticsearchConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
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
    format            = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
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
`, name, domain)
}
