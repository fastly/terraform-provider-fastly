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

func TestResourceFastlyFlattenDigitalOcean(t *testing.T) {
	cases := []struct {
		remote []*gofastly.DigitalOcean
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.DigitalOcean{
				{
					ServiceVersion:    1,
					Name:              "digitalocean-endpoint",
					BucketName:        "bucket",
					AccessKey:         "access",
					SecretKey:         "secret",
					Domain:            "nyc3.digitaloceanspaces.com",
					PublicKey:         pgpPublicKey(t),
					Path:              "/",
					Period:            3600,
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
					GzipLevel:         1,
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     2,
					MessageType:       "classic",
					Placement:         "none",
					ResponseCondition: "always",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "digitalocean-endpoint",
					"bucket_name":        "bucket",
					"access_key":         "access",
					"secret_key":         "secret",
					"domain":             "nyc3.digitaloceanspaces.com",
					"public_key":         pgpPublicKey(t),
					"path":               "/",
					"period":             uint(3600),
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"gzip_level":         uint(1),
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(2),
					"message_type":       "classic",
					"placement":          "none",
					"response_condition": "always",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDigitalOcean(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceV1_logging_digitalocean_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.DigitalOcean{
		ServiceVersion:    1,
		Name:              "digitalocean-endpoint",
		BucketName:        "bucket",
		AccessKey:         "access",
		SecretKey:         "secret",
		Domain:            "nyc3.digitaloceanspaces.com",
		PublicKey:         pgpPublicKey(t),
		Path:              "/",
		Period:            3600,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         0,
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     2,
		MessageType:       "classic",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
	}

	log1_after_update := gofastly.DigitalOcean{
		ServiceVersion:    1,
		Name:              "digitalocean-endpoint",
		BucketName:        "bucketupdate",
		AccessKey:         "accessupdate",
		SecretKey:         "secretupdate",
		Domain:            "nyc4.digitaloceanspaces.com",
		PublicKey:         pgpPublicKey(t),
		Path:              "new/",
		Period:            3601,
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         2,
		FormatVersion:     2,
		MessageType:       "blank",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
	}

	log2 := gofastly.DigitalOcean{
		ServiceVersion:    1,
		Name:              "another-digitalocean-endpoint",
		BucketName:        "bucket2",
		AccessKey:         "access2",
		SecretKey:         "secret2",
		Domain:            "nyc3.digitaloceanspaces.com",
		PublicKey:         pgpPublicKey(t),
		Path:              "two/",
		Period:            3600,
		Format:            "%h %l %u %t \"%r\" %>s %b",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:         1,
		FormatVersion:     2,
		MessageType:       "classic",
		Placement:         "none",
		ResponseCondition: "response_condition_test",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DigitalOceanConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_digitalocean.#", "1"),
				),
			},

			{
				Config: testAccServiceV1DigitalOceanConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_digitalocean.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_digitalocean_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.DigitalOcean{
		ServiceVersion:  1,
		Name:            "digitalocean-endpoint",
		BucketName:      "bucket",
		AccessKey:       "access",
		SecretKey:       "secret",
		Domain:          "nyc3.digitaloceanspaces.com",
		PublicKey:       pgpPublicKey(t),
		Path:            "/",
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
		GzipLevel:       0,
		MessageType:     "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DigitalOceanComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1DigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_digitalocean.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DigitalOceanAttributes(service *gofastly.ServiceDetail, digitalocean []*gofastly.DigitalOcean, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		digitaloceanList, err := conn.ListDigitalOceans(&gofastly.ListDigitalOceansInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up DigitalOcean Spaces Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(digitaloceanList) != len(digitalocean) {
			return fmt.Errorf("DigitalOcean Spaces List count mismatch, expected (%d), got (%d)", len(digitalocean), len(digitaloceanList))
		}

		log.Printf("[DEBUG] digitaloceanList = %#v\n", digitaloceanList)

		for _, e := range digitalocean {
			for _, el := range digitaloceanList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
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
						return fmt.Errorf("Bad match DigitalOcean Spaces logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceV1DigitalOceanConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-digitalocean-logging"
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

  logging_digitalocean {
    name   = "digitalocean-endpoint"
    bucket_name = "bucket"
    access_key = "access"
    secret_key = "secret"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level = 0
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    placement = "none"
    response_condition = "response_condition_test"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1DigitalOceanConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-digitalocean-logging"
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

  logging_digitalocean {
    name   = "digitalocean-endpoint"
    bucket_name = "bucketupdate"
    access_key = "accessupdate"
    secret_key = "secretupdate"
    domain = "nyc4.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "new/"
    period = 3601
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level = 2
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    message_type = "blank"
    placement = "none"
    response_condition = "response_condition_test"
  }

  logging_digitalocean {
    name   = "another-digitalocean-endpoint"
    bucket_name = "bucket2"
    access_key = "access2"
    secret_key = "secret2"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "two/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level = 1
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    placement = "none"
    response_condition = "response_condition_test"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1DigitalOceanComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-digitalocean-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_digitalocean {
    name   = "digitalocean-endpoint"
    bucket_name = "bucket"
    access_key = "access"
    secret_key = "secret"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    gzip_level = 0
    message_type = "classic"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}
