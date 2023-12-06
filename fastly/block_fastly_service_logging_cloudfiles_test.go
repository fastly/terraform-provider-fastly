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

func TestResourceFastlyFlattenCloudfiles(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Cloudfiles
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Cloudfiles{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("cloudfiles-endpoint"),
					BucketName:        gofastly.ToPointer("bucket"),
					User:              gofastly.ToPointer("user"),
					AccessKey:         gofastly.ToPointer("secret"),
					PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					FormatVersion:     gofastly.ToPointer(2),
					GzipLevel:         gofastly.ToPointer(0),
					MessageType:       gofastly.ToPointer("classic"),
					Path:              gofastly.ToPointer("/"),
					Region:            gofastly.ToPointer("ORD"),
					Period:            gofastly.ToPointer(3600),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("response_condition"),
					TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
					CompressionCodec:  gofastly.ToPointer("zstd"),
				},
			},
			local: []map[string]any{
				{
					"name":               "cloudfiles-endpoint",
					"bucket_name":        "bucket",
					"user":               "user",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             "%h %l %u %t \"%r\" %>s %b",
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

func TestAccFastlyServiceVCL_logging_cloudfiles_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Cloudfiles{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("cloudfiles-endpoint"),
		BucketName:        gofastly.ToPointer("bucket"),
		User:              gofastly.ToPointer("user"),
		AccessKey:         gofastly.ToPointer("secret"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Path:              gofastly.ToPointer("/"),
		Region:            gofastly.ToPointer("ORD"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
	}

	log1AfterUpdate := gofastly.Cloudfiles{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("cloudfiles-endpoint"),
		BucketName:        gofastly.ToPointer("bucketupdate"),
		User:              gofastly.ToPointer("userupdate"),
		AccessKey:         gofastly.ToPointer("secretupdate"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(1),
		MessageType:       gofastly.ToPointer("blank"),
		Path:              gofastly.ToPointer("new/"),
		Region:            gofastly.ToPointer("LON"),
		Period:            gofastly.ToPointer(3601),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		CompressionCodec:  gofastly.ToPointer(""),
	}

	log2 := gofastly.Cloudfiles{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("another-cloudfiles-endpoint"),
		BucketName:        gofastly.ToPointer("bucket2"),
		User:              gofastly.ToPointer("user2"),
		AccessKey:         gofastly.ToPointer("secret2"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Path:              gofastly.ToPointer("two/"),
		Region:            gofastly.ToPointer("SYD"),
		Period:            gofastly.ToPointer(3600),
		Placement:         gofastly.ToPointer("none"),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
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
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.none", "logging_cloudfiles.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLCloudfilesConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.none", &service),
					testAccCheckFastlyServiceVCLCloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.none", "logging_cloudfiles.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_cloudfiles_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1Compute := gofastly.Cloudfiles{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("cloudfiles-endpoint"),
		BucketName:       gofastly.ToPointer("bucket"),
		User:             gofastly.ToPointer("user"),
		AccessKey:        gofastly.ToPointer("secret"),
		PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
		GzipLevel:        gofastly.ToPointer(0),
		MessageType:      gofastly.ToPointer("classic"),
		Path:             gofastly.ToPointer("/"),
		Region:           gofastly.ToPointer("ORD"),
		Period:           gofastly.ToPointer(3600),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		CompressionCodec: gofastly.ToPointer("zstd"),
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
		cloudfilesList, err := conn.ListCloudfiles(&gofastly.ListCloudfilesInput{
			ServiceID:      gofastly.ToValue(service.ID),
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
					e.ServiceID = service.ID
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
    name = "cloudfiles-endpoint"
    bucket_name = "bucket"
    user = "user"
    access_key = "secret"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    path = "/"
    region = "ORD"
    period = 3600
    placement = "none"
    response_condition = "response_condition_test"
    format_version = 2
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    compression_codec = "zstd"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLCloudfilesConfigUpdate(name, domain string) string {
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
    name = "cloudfiles-endpoint"
    bucket_name = "bucketupdate"
    user = "userupdate"
    access_key = "secretupdate"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    message_type = "blank"
    path = "new/"
    region = "LON"
    period = 3601
    placement = "none"
    response_condition = "response_condition_test"
    gzip_level = 1
    format_version = 2
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
  }

  logging_cloudfiles {
    name = "another-cloudfiles-endpoint"
    bucket_name = "bucket2"
    user = "user2"
    access_key = "secret2"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    message_type = "classic"
    path = "two/"
    region = "SYD"
    period = 3600
    placement = "none"
    response_condition = "response_condition_test"
    format_version = 2
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    compression_codec = "zstd"
  }

  force_destroy = true
}
`, name, domain)
}
