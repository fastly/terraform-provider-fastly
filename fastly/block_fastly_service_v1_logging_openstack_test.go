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

func TestResourceFastlyFlattenOpenstack(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Openstack
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Openstack{
				{
					Name:              "openstack-logging",
					URL:               "https://auth.example.com",
					User:              "user",
					BucketName:        "bucket",
					AccessKey:         "secret",
					PublicKey:         pgpPublicKey(t),
					Format:            "log format",
					FormatVersion:     2,
					MessageType:       "classic",
					Path:              "/",
					Placement:         "none",
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
					ResponseCondition: "always",
					Period:            3600,
					GzipLevel:         0,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "openstack-logging",
					"url":                "https://auth.example.com",
					"user":               "user",
					"bucket_name":        "bucket",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             "log format",
					"format_version":     uint(2),
					"message_type":       "classic",
					"path":               "/",
					"placement":          "none",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"response_condition": "always",
					"period":             uint(3600),
					"gzip_level":         uint(0),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenOpenstack(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceV1_logging_openstack_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Openstack{
		Version:           1,
		Name:              "openstack-endpoint",
		URL:               "https://auth.example.com/v1", // /v1, /v2 or /v3 are required to be in the path.
		User:              "user",
		BucketName:        "bucket",
		AccessKey:         "s3cr3t",
		PublicKey:         pgpPublicKey(t),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     2,
		MessageType:       "classic",
		Path:              "/",
		Placement:         "none",
		TimestampFormat:   `%Y-%m-%dT%H:%M:%S.000`,
		ResponseCondition: "response_condition_test",
		Period:            3600,
		GzipLevel:         0,
	}

	log1_after_update := gofastly.Openstack{
		Version:           1,
		Name:              "openstack-endpoint",
		Format:            "%h %l %u %t \"%r\" %>s %b %T",
		URL:               "https://auth.example.com/v2", // /v1, /v2 or /v3 are required to be in the path.
		User:              "userupdate",
		BucketName:        "bucketupdate",
		AccessKey:         "s3cr3tupdate",
		PublicKey:         pgpPublicKey(t),
		FormatVersion:     2,
		MessageType:       "blank",
		Path:              "new/",
		Placement:         "none",
		TimestampFormat:   `%Y-%m-%dT%H:%M:%S.000`,
		ResponseCondition: "response_condition_test",
		Period:            3601,
		GzipLevel:         1,
	}

	log2 := gofastly.Openstack{
		Version:           1,
		Name:              "another-openstack-endpoint",
		URL:               "https://auth.example.com/v3", // /v1, /v2 or /v3 are required to be in the path.
		User:              "user2",
		BucketName:        "bucket2",
		AccessKey:         "s3cr3t2",
		PublicKey:         pgpPublicKey(t),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     2,
		MessageType:       "classic",
		Path:              "two/",
		Placement:         "none",
		TimestampFormat:   `%Y-%m-%dT%H:%M:%S.000`,
		ResponseCondition: "response_condition_test",
		Period:            3600,
		GzipLevel:         0,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1OpenstackConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1OpenstackAttributes(&service, []*gofastly.Openstack{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_openstack.#", "1"),
				),
			},

			{
				Config: testAccServiceV1OpenstackConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1OpenstackAttributes(&service, []*gofastly.Openstack{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_openstack.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_logging_openstack_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Openstack{
		Version:         1,
		Name:            "openstack-endpoint",
		URL:             "https://auth.example.com/v1", // /v1, /v2 or /v3 are required to be in the path.
		User:            "user",
		BucketName:      "bucket",
		AccessKey:       "s3cr3t",
		PublicKey:       pgpPublicKey(t),
		MessageType:     "classic",
		Path:            "/",
		TimestampFormat: `%Y-%m-%dT%H:%M:%S.000`,
		Period:          3600,
		GzipLevel:       0,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1OpenstackComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1OpenstackAttributes(&service, []*gofastly.Openstack{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_openstack.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1OpenstackAttributes(service *gofastly.ServiceDetail, openstack []*gofastly.Openstack, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		openstackList, err := conn.ListOpenstack(&gofastly.ListOpenstackInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up OpenStack Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(openstackList) != len(openstack) {
			return fmt.Errorf("OpenStack List count mismatch, expected (%d), got (%d)", len(openstack), len(openstackList))
		}

		log.Printf("[DEBUG] openstackList = %#v\n", openstackList)

		for _, e := range openstack {
			for _, el := range openstackList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.Version = service.ActiveVersion.Number
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
						return fmt.Errorf("Bad match OpenStack logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceV1OpenstackConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    path = "/"
    placement = "none"
		timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1OpenstackConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
    user   = "userupdate"
		url    = "https://auth.example.com/v2"
    bucket_name = "bucketupdate"
    access_key = "s3cr3tupdate"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
    path = "new/"
    placement = "none"
		timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
		message_type = "blank"
		gzip_level = 1
    period = 3601
  }

  logging_openstack {
    name   = "another-openstack-endpoint"
		url    = "https://auth.example.com/v3"
    user   = "user2"
    bucket_name = "bucket2"
    access_key = "s3cr3t2"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    path = "two/"
    placement = "none"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    response_condition = "response_condition_test"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceV1OpenstackComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-openstack-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_openstack {
    name   = "openstack-endpoint"
    url    = "https://auth.example.com/v1"
    user   = "user"
    bucket_name = "bucket"
    access_key = "s3cr3t"
    public_key = file("test_fixtures/fastly_test_publickey")
    path = "/"
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
