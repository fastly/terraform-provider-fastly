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

func TestAccFastlyServiceLoggingOpenStack_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Openstack{
		AccessKey:         new("s3cr3t"),
		BucketName:        new("bucket"),
		CompressionCodec:  new("zstd"),
		Format:            new(LoggingOpenStackDefaultFormat),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		MessageType:       new("classic"),
		Name:              new("openstack-endpoint"),
		Path:              new("/"),
		Period:            new(3600),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new(`%Y-%m-%dT%H:%M:%S.000`),
		URL:               new("https://auth.example.com/v1"), // /v1, /v2 or /v3 are required to be in the path.
		User:              new("user"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Openstack{
		AccessKey:         new("s3cr3tupdate"),
		BucketName:        new("bucketupdate"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(1),
		MessageType:       new("blank"),
		Name:              new("openstack-endpoint"),
		Path:              new("new/"),
		Period:            new(3601),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new(`%Y-%m-%dT%H:%M:%S.000`),
		URL:               new("https://auth.example.com/v2"), // /v1, /v2 or /v3 are required to be in the path.
		User:              new("userupdate"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Openstack{
		AccessKey:         new("s3cr3t2"),
		BucketName:        new("bucket2"),
		CompressionCodec:  new("zstd"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		MessageType:       new("classic"),
		Name:              new("another-openstack-endpoint"),
		Path:              new("two/"),
		Period:            new(3600),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new(`%Y-%m-%dT%H:%M:%S.000`),
		URL:               new("https://auth.example.com/v3"), // /v1, /v2 or /v3 are required to be in the path.
		User:              new("user2"),
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
				Config: testAccServiceVCLOpenstackConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLOpenstackAttributes(&service, []*gofastly.Openstack{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_openstack.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLOpenstackConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLOpenstackAttributes(&service, []*gofastly.Openstack{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_openstack.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingOpenStack_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Openstack{
		AccessKey:        new("s3cr3t"),
		BucketName:       new("bucket"),
		CompressionCodec: new("zstd"),
		GzipLevel:        new(0),
		MessageType:      new("classic"),
		Name:             new("openstack-endpoint"),
		Path:             new("/"),
		Period:           new(3600),
		PublicKey:        new(pgpPublicKey(t)),
		ServiceVersion:   new(1),
		TimestampFormat:  new(`%Y-%m-%dT%H:%M:%S.000`),
		URL:              new("https://auth.example.com/v1"), // /v1, /v2 or /v3 are required to be in the path.
		User:             new("user"),
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
				Config: testAccServiceVCLOpenstackComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLOpenstackAttributes(&service, []*gofastly.Openstack{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_openstack.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLOpenstackAttributes(service *gofastly.ServiceDetail, openstack []*gofastly.Openstack, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		openstackList, err := conn.ListOpenstack(context.TODO(), &gofastly.ListOpenstackInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up OpenStack Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(openstackList) != len(openstack) {
			return fmt.Errorf("openStack List count mismatch, expected (%d), got (%d)", len(openstack), len(openstackList))
		}

		log.Printf("[DEBUG] openstackList = %#v\n", openstackList)

		for _, e := range openstack {
			for _, el := range openstackList {
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
						return fmt.Errorf("bad match OpenStack logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLOpenstackConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-openstack-logging"
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

  logging_openstack {
    name   = "openstack-endpoint"
	url    = "https://auth.example.com/v1"
    user   = "user"
    bucket_name = "bucket"
    access_key = "s3cr3t"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    placement = "none"
	timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
    compression_codec = "zstd"
    processing_region = "us"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLOpenstackConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-openstack-logging"
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

  logging_openstack {
    name = "openstack-endpoint"
    user = "userupdate"
    url = "https://auth.example.com/v2"
    bucket_name = "bucketupdate"
    access_key = "s3cr3tupdate"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = %q
    path = "new/"
    placement = "none"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
    message_type = "blank"
    gzip_level = 1
    period = 3601
  }

  logging_openstack {
    name = "another-openstack-endpoint"
    url = "https://auth.example.com/v3"
    user = "user2"
    bucket_name = "bucket2"
    access_key = "s3cr3t2"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = %q
    path = "two/"
    placement = "none"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain, format, format)
}

func testAccServiceVCLOpenstackComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-openstack-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_openstack {
    name = "openstack-endpoint"
    url = "https://auth.example.com/v1"
    user = "user"
    bucket_name = "bucket"
    access_key = "s3cr3t"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    compression_codec = "zstd"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domain)
}

func TestResourceFastlyFlattenOpenstack(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Openstack
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Openstack{
				{
					Name:              new("openstack-logging"),
					URL:               new("https://auth.example.com"),
					User:              new("user"),
					BucketName:        new("bucket"),
					AccessKey:         new("secret"),
					PublicKey:         new(pgpPublicKey(t)),
					Format:            new(LoggingOpenStackDefaultFormat),
					FormatVersion:     new(2),
					MessageType:       new("classic"),
					Path:              new("/"),
					Placement:         new("none"),
					TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
					ResponseCondition: new("always"),
					Period:            new(3600),
					GzipLevel:         new(0),
					CompressionCodec:  new("zstd"),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "openstack-logging",
					"url":                "https://auth.example.com",
					"user":               "user",
					"bucket_name":        "bucket",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             LoggingOpenStackDefaultFormat,
					"format_version":     2,
					"message_type":       "classic",
					"path":               "/",
					"placement":          "none",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"response_condition": "always",
					"period":             3600,
					"gzip_level":         0,
					"compression_codec":  "zstd",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenOpenstack(c.remote, nil)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
