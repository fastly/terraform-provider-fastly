package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func TestAccFastlyServiceLoggingCloudfiles_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Cloudfiles{
		AccessKey:         new("secret"),
		BucketName:        new("bucket"),
		CompressionCodec:  new("zstd"),
		Format:            new(LoggingCloudFilesDefaultFormat),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		MessageType:       new("classic"),
		Name:              new("cloudfiles-endpoint"),
		Path:              new("/"),
		Period:            new(3600),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		Region:            new("ORD"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
		User:              new("user"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Cloudfiles{
		AccessKey:         new("secretupdate"),
		BucketName:        new("bucketupdate"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(1),
		MessageType:       new("blank"),
		Name:              new("cloudfiles-endpoint"),
		Path:              new("new/"),
		Period:            new(3601),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		Region:            new("LON"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
		User:              new("userupdate"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Cloudfiles{
		AccessKey:         new("secret2"),
		BucketName:        new("bucket2"),
		CompressionCodec:  new("zstd"),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		MessageType:       new("classic"),
		Name:              new("another-cloudfiles-endpoint"),
		Path:              new("two/"),
		Period:            new(3600),
		Placement:         new("none"),
		PublicKey:         new(pgpPublicKey(t)),
		Region:            new("SYD"),
		ResponseCondition: new("response_condition_test"),
		ServiceVersion:    new(1),
		TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
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
				Config: testAccServiceVCLCloudfilesConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.none", &service),
					testAccCheckFastlyServiceVCLCloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.none", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.none", "logging_cloudfiles.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLCloudfilesConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.none", &service),
					testAccCheckFastlyServiceVCLCloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.none", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.none", "logging_cloudfiles.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingCloudfiles_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1Compute := gofastly.Cloudfiles{
		ServiceVersion:   new(1),
		Name:             new("cloudfiles-endpoint"),
		BucketName:       new("bucket"),
		User:             new("user"),
		AccessKey:        new("secret"),
		PublicKey:        new(pgpPublicKey(t)),
		GzipLevel:        new(0),
		MessageType:      new("classic"),
		Path:             new("/"),
		Region:           new("ORD"),
		Period:           new(3600),
		TimestampFormat:  new("%Y-%m-%dT%H:%M:%S.000"),
		CompressionCodec: new("zstd"),
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
				Config: testAccServiceVCLComputeCloudfilesConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.none", &service),
					testAccCheckFastlyServiceVCLCloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1Compute}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.none", "logging_cloudfiles.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLCloudfilesAttributes(service *gofastly.ServiceDetail, cloudfiles []*gofastly.Cloudfiles, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		cloudfilesList, err := conn.ListCloudfiles(context.TODO(), &gofastly.ListCloudfilesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Cloud Files Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(cloudfilesList) != len(cloudfiles) {
			return fmt.Errorf("cloud Files List count mismatch, expected (%d), got (%d)", len(cloudfiles), len(cloudfilesList))
		}

		log.Printf("[DEBUG] cloudfilesList = %#v\n", cloudfilesList)

		for _, e := range cloudfiles {
			for _, el := range cloudfilesList {
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
						return fmt.Errorf("bad match Cloud Files logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLComputeCloudfilesConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "none" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-cloudfiles-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_cloudfiles {
    name = "cloudfiles-endpoint"
    bucket_name = "bucket"
    user = "user"
    access_key = "secret"
    public_key = file("test_fixtures/fastly_test_publickey")
    message_type = "classic"
    path = "/"
    region = "ORD"
    period = 3600
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    compression_codec = "zstd"
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

func testAccServiceVCLCloudfilesConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "none" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-cloudfiles-logging"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

  logging_cloudfiles {
    access_key = "secret"
    bucket_name = "bucket"
    compression_codec = "zstd"
    format_version = 2
    message_type = "classic"
    name = "cloudfiles-endpoint"
    path = "/"
    period = 3600
    placement = "none"
    public_key = file("test_fixtures/fastly_test_publickey")
    region = "ORD"
    response_condition = "response_condition_test"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    user = "user"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLCloudfilesConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "none" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-cloudfiles-logging"
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

  logging_cloudfiles {
    access_key = "secretupdate"
    bucket_name = "bucketupdate"
    format = %q
    format_version = 2
    gzip_level = 1
    message_type = "blank"
    name = "cloudfiles-endpoint"
    path = "new/"
    period = 3601
    placement = "none"
    public_key = file("test_fixtures/fastly_test_publickey")
    region = "LON"
    response_condition = "response_condition_test"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    user = "userupdate"
  }

  logging_cloudfiles {
    access_key = "secret2"
    bucket_name = "bucket2"
    compression_codec = "zstd"
    format = %q
    format_version = 2
    message_type = "classic"
    name = "another-cloudfiles-endpoint"
    path = "two/"
    period = 3600
    placement = "none"
    public_key = file("test_fixtures/fastly_test_publickey")
    region = "SYD"
    response_condition = "response_condition_test"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    user = "user2"
  }

  force_destroy = true
}
`, name, domain, strings.ReplaceAll(format, "%{", "%%{"), strings.ReplaceAll(format, "%{", "%%{"))
}

func TestResourceFastlyFlattenCloudfiles(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Cloudfiles
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Cloudfiles{
				{
					ServiceVersion:    new(1),
					Name:              new("cloudfiles-endpoint"),
					BucketName:        new("bucket"),
					User:              new("user"),
					AccessKey:         new("secret"),
					PublicKey:         new(pgpPublicKey(t)),
					Format:            new(LoggingCloudFilesDefaultFormat),
					FormatVersion:     new(2),
					GzipLevel:         new(0),
					MessageType:       new("classic"),
					Path:              new("/"),
					Region:            new("ORD"),
					Period:            new(3600),
					Placement:         new("none"),
					ResponseCondition: new("response_condition"),
					TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
					CompressionCodec:  new("zstd"),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "cloudfiles-endpoint",
					"bucket_name":        "bucket",
					"user":               "user",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             LoggingCloudFilesDefaultFormat,
					"format_version":     2,
					"gzip_level":         0,
					"message_type":       "classic",
					"path":               "/",
					"region":             "ORD",
					"period":             3600,
					"placement":          "none",
					"response_condition": "response_condition",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"compression_codec":  "zstd",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenCloudfiles(c.remote, nil)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
