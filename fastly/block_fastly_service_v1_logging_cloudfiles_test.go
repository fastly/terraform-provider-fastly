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

func TestResourceFastlyFlattenCloudfiles(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Cloudfiles
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Cloudfiles{
				{
					Version:           1,
					Name:              "cloudfiles-endpoint",
					BucketName:        "bucket",
					User:              "user",
					AccessKey:         "secret",
					PublicKey:         pgpPublicKey(t),
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     2,
					GzipLevel:         0,
					MessageType:       "classic",
					Path:              "/",
					Region:            "ORD",
					Period:            3600,
					Placement:         "none",
					ResponseCondition: "response_condition",
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "cloudfiles-endpoint",
					"bucket_name":        "bucket",
					"user":               "user",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(2),
					"gzip_level":         uint(0),
					"message_type":       "classic",
					"path":               "/",
					"region":             "ORD",
					"period":             uint(3600),
					"placement":          "none",
					"response_condition": "response_condition",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenCloudfiles(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceV1_logging_cloudfiles_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Cloudfiles{
		Version:           1,
		Name:              "cloudfiles-endpoint",
		BucketName:        "bucket",
		User:              "user",
		AccessKey:         "secret",
		PublicKey:         pgpPublicKey(t),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     2,
		GzipLevel:         0,
		MessageType:       "classic",
		Path:              "/",
		Region:            "ORD",
		Period:            3600,
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
	}

	log1_after_update := gofastly.Cloudfiles{
		Version:           1,
		Name:              "cloudfiles-endpoint",
		BucketName:        "bucketupdate",
		User:              "userupdate",
		AccessKey:         "secretupdate",
		PublicKey:         pgpPublicKey(t),
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
		FormatVersion:     2,
		GzipLevel:         1,
		MessageType:       "blank",
		Path:              "new/",
		Region:            "LON",
		Period:            3601,
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
	}

	log2 := gofastly.Cloudfiles{
		Version:           1,
		Name:              "another-cloudfiles-endpoint",
		BucketName:        "bucket2",
		User:              "user2",
		AccessKey:         "secret2",
		PublicKey:         pgpPublicKey(t),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     2,
		GzipLevel:         0,
		MessageType:       "classic",
		Path:              "two/",
		Region:            "SYD",
		Period:            3600,
		Placement:         "none",
		ResponseCondition: "response_condition_test",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1CloudfilesConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.none", &service),
					testAccCheckFastlyServiceV1CloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.none", "logging_cloudfiles.#", "1"),
				),
			},

			{
				Config: testAccServiceV1CloudfilesConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.none", &service),
					testAccCheckFastlyServiceV1CloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.none", "logging_cloudfiles.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_cloudfiles_basicWasm(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1Wasm := gofastly.Cloudfiles{
		Version:         1,
		Name:            "cloudfiles-endpoint",
		BucketName:      "bucket",
		User:            "user",
		AccessKey:       "secret",
		PublicKey:       pgpPublicKey(t),
		GzipLevel:       0,
		MessageType:     "classic",
		Path:            "/",
		Region:          "ORD",
		Period:          3600,
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WasmCloudfilesConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.none", &service),
					testAccCheckFastlyServiceV1CloudfilesAttributes(&service, []*gofastly.Cloudfiles{&log1Wasm}, ServiceTypeWasm),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.none", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.none", "logging_cloudfiles.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1CloudfilesAttributes(service *gofastly.ServiceDetail, cloudfiles []*gofastly.Cloudfiles, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		cloudfilesList, err := conn.ListCloudfiles(&gofastly.ListCloudfilesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Cloud Files Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(cloudfilesList) != len(cloudfiles) {
			return fmt.Errorf("Cloud Files List count mismatch, expected (%d), got (%d)", len(cloudfiles), len(cloudfilesList))
		}

		log.Printf("[DEBUG] cloudfilesList = %#v\n", cloudfilesList)

		for _, e := range cloudfiles {
			for _, el := range cloudfilesList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					el.CreatedAt = nil
					el.UpdatedAt = nil

					// Ignore VCL attributes for Wasm and set to whatever is returned from the API.
					if serviceType == ServiceTypeWasm {
						el.FormatVersion = e.FormatVersion
						el.Format = e.Format
						el.ResponseCondition = e.ResponseCondition
						el.Placement = e.Placement
					}

					if diff := cmp.Diff(e, el); diff != "" {
						return fmt.Errorf("Bad match Cloud Files logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceV1WasmCloudfilesConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "none" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-cloudfiles-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_cloudfiles {
    name   = "cloudfiles-endpoint"
    bucket_name = "bucket"
    user = "user"
    access_key = "secret"
    public_key = file("test_fixtures/fastly_test_publickey")
    message_type = "classic"
	path = "/"
	region = "ORD"
	period = 3600
	gzip_level = 0
	timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
  }
 
  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1CloudfilesConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "none" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-cloudfiles-logging"
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

  logging_cloudfiles {
    name   = "cloudfiles-endpoint"
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
    gzip_level = 0
    format_version = 2
		timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1CloudfilesConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "none" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-cloudfiles-logging"
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

  logging_cloudfiles {
    name   = "cloudfiles-endpoint"
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
    name   = "another-cloudfiles-endpoint"
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
    gzip_level = 0
    format_version = 2
		timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
  }

  force_destroy = true
}
`, name, domain)
}
