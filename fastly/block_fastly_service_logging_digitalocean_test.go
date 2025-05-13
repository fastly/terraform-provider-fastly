package fastly

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

func TestResourceFastlyFlattenDigitalOcean(t *testing.T) {
	cases := []struct {
		remote []*gofastly.DigitalOcean
		local  []map[string]any
	}{
		{
			remote: []*gofastly.DigitalOcean{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("digitalocean-endpoint"),
					BucketName:        gofastly.ToPointer("bucket"),
					AccessKey:         gofastly.ToPointer("access"),
					SecretKey:         gofastly.ToPointer("secret"),
					Domain:            gofastly.ToPointer("nyc3.digitaloceanspaces.com"),
					PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
					Path:              gofastly.ToPointer("/"),
					Period:            gofastly.ToPointer(3600),
					TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
					GzipLevel:         gofastly.ToPointer(0),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					FormatVersion:     gofastly.ToPointer(2),
					MessageType:       gofastly.ToPointer("classic"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					CompressionCodec:  gofastly.ToPointer("zstd"),
				},
			},
			local: []map[string]any{
				{
					"name":               "digitalocean-endpoint",
					"bucket_name":        "bucket",
					"access_key":         "access",
					"secret_key":         "secret",
					"domain":             "nyc3.digitaloceanspaces.com",
					"public_key":         pgpPublicKey(t),
					"path":               "/",
					"period":             3600,
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"gzip_level":         0,
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     2,
					"message_type":       "classic",
					"placement":          "none",
					"response_condition": "always",
					"compression_codec":  "zstd",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDigitalOcean(c.remote, nil)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCL_logging_digitalocean_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.DigitalOcean{
		AccessKey:         gofastly.ToPointer("access"),
		BucketName:        gofastly.ToPointer("bucket"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Domain:            gofastly.ToPointer("nyc3.digitaloceanspaces.com"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("digitalocean-endpoint"),
		Path:              gofastly.ToPointer("/"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SecretKey:         gofastly.ToPointer("secret"),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
	}

	log1AfterUpdate := gofastly.DigitalOcean{
		AccessKey:         gofastly.ToPointer("accessupdate"),
		BucketName:        gofastly.ToPointer("bucketupdate"),
		Domain:            gofastly.ToPointer("nyc4.digitaloceanspaces.com"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(2),
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("digitalocean-endpoint"),
		Path:              gofastly.ToPointer("new/"),
		Period:            gofastly.ToPointer(3601),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SecretKey:         gofastly.ToPointer("secretupdate"),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
	}

	log2 := gofastly.DigitalOcean{
		AccessKey:         gofastly.ToPointer("access2"),
		BucketName:        gofastly.ToPointer("bucket2"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Domain:            gofastly.ToPointer("nyc3.digitaloceanspaces.com"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("another-digitalocean-endpoint"),
		Path:              gofastly.ToPointer("two/"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		SecretKey:         gofastly.ToPointer("secret2"),
		ServiceVersion:    gofastly.ToPointer(1),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLDigitalOceanConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_digitalocean.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLDigitalOceanConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_digitalocean.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_digitalocean_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.DigitalOcean{
		AccessKey:        gofastly.ToPointer("access"),
		BucketName:       gofastly.ToPointer("bucket"),
		CompressionCodec: gofastly.ToPointer("zstd"),
		Domain:           gofastly.ToPointer("nyc3.digitaloceanspaces.com"),
		GzipLevel:        gofastly.ToPointer(0),
		MessageType:      gofastly.ToPointer("classic"),
		Name:             gofastly.ToPointer("digitalocean-endpoint"),
		Path:             gofastly.ToPointer("/"),
		Period:           gofastly.ToPointer(3600),
		PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
		SecretKey:        gofastly.ToPointer("secret"),
		ServiceVersion:   gofastly.ToPointer(1),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLDigitalOceanComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLDigitalOceanAttributes(&service, []*gofastly.DigitalOcean{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_digitalocean.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLDigitalOceanAttributes(service *gofastly.ServiceDetail, digitalocean []*gofastly.DigitalOcean, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		digitaloceanList, err := conn.ListDigitalOceans(&gofastly.ListDigitalOceansInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up DigitalOcean Spaces Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(digitaloceanList) != len(digitalocean) {
			return fmt.Errorf("digitalOcean Spaces List count mismatch, expected (%d), got (%d)", len(digitalocean), len(digitaloceanList))
		}

		log.Printf("[DEBUG] digitaloceanList = %#v\n", digitaloceanList)

		for _, e := range digitalocean {
			for _, el := range digitaloceanList {
				if gofastly.ToValue(e.Name) == gofastly.ToValue(el.Name) {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ServiceID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
						return fmt.Errorf("bad match DigitalOcean Spaces logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLDigitalOceanConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-digitalocean-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_digitalocean {
    name = "digitalocean-endpoint"
    bucket_name = "bucket"
    access_key = "access"
    secret_key = "secret"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    placement = "none"
    response_condition = "response_condition_test"
    compression_codec = "zstd"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLDigitalOceanConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-digitalocean-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_digitalocean {
    name = "digitalocean-endpoint"
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
    name = "another-digitalocean-endpoint"
    bucket_name = "bucket2"
    access_key = "access2"
    secret_key = "secret2"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "two/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    placement = "none"
    response_condition = "response_condition_test"
    compression_codec = "zstd"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLDigitalOceanComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-digitalocean-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_digitalocean {
    name = "digitalocean-endpoint"
    bucket_name = "bucket"
    access_key = "access"
    secret_key = "secret"
    domain = "nyc3.digitaloceanspaces.com"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    message_type = "classic"
    compression_codec = "zstd"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
