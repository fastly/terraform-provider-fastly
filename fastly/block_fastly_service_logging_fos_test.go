package fastly

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
)

func TestAccFastlyServiceLoggingFOS_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.FOS{
		BucketName:        new("fastlytestlogging"),
		CompressionCodec:  new("zstd"),
		FileMaxBytes:      new(1048576),
		Format:            new(LoggingFOSDefaultFormat),
		FormatVersion:     new(2),
		GzipLevel:         new(0),
		MessageType:       new("classic"),
		Name:              new("somefoslog"),
		Path:              new(""),
		Period:            new(3600),
		PublicKey:         new(pgpPublicKey(t)),
		Region:            new("us-east"),
		ResponseCondition: new(""),
		ServiceVersion:    new(1),
		TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLFOSLoggingConfig(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceFOSLoggingAttributes(&service, []*gofastly.FOS{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_fos.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingFOS_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.FOS{
		BucketName:       new("fastlytestlogging"),
		CompressionCodec: new("zstd"),
		FileMaxBytes:     new(1048576),
		GzipLevel:        new(0),
		MessageType:      new("classic"),
		Name:             new("somefoslog"),
		Path:             new(""),
		Period:           new(3600),
		PublicKey:        new(pgpPublicKey(t)),
		Region:           new("us-east"),
		ServiceVersion:   new(1),
		TimestampFormat:  new("%Y-%m-%dT%H:%M:%S.000"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeFOSLoggingConfig(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceFOSLoggingAttributes(&service, []*gofastly.FOS{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_fos.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceFOSLoggingAttributes(service *gofastly.ServiceDetail, foses []*gofastly.FOS, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		remote, err := conn.ListFOSes(context.TODO(), &gofastly.ListFOSesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up FOS logging for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(remote) != len(foses) {
			return fmt.Errorf("FOS list count mismatch, expected (%d), got (%d)", len(foses), len(remote))
		}

		var found int
		for _, f := range foses {
			for _, rf := range remote {
				if gofastly.ToValue(f.Name) == gofastly.ToValue(rf.Name) {
					// Not known ahead of time.
					f.ServiceID = service.ServiceID
					f.ServiceVersion = service.ActiveVersion.Number
					// Server-generated; not tracked.
					rf.AccessKey = nil
					rf.CreatedAt = nil
					rf.UpdatedAt = nil

					// VCL-only attributes don't apply to Compute.
					if serviceType == ServiceTypeCompute {
						rf.FormatVersion = f.FormatVersion
						rf.Format = f.Format
						rf.ResponseCondition = f.ResponseCondition
						rf.Placement = f.Placement
					}

					if diff := cmp.Diff(f, rf); diff != "" {
						return fmt.Errorf("bad match FOS logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(foses) {
			return fmt.Errorf("error matching FOS logging rules")
		}

		return nil
	}
}

func testAccServiceVCLFOSLoggingConfig(name, domain string) string {
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

  logging_fos {
    name              = "somefoslog"
    bucket_name       = "fastlytestlogging"
    region            = "us-east"
    file_max_bytes    = 1048576
    public_key        = file("test_fixtures/fastly_test_publickey")
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceComputeFOSLoggingConfig(name, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_fos {
    name              = "somefoslog"
    bucket_name       = "fastlytestlogging"
    region            = "us-east"
    file_max_bytes    = 1048576
    public_key        = file("test_fixtures/fastly_test_publickey")
    compression_codec = "zstd"
  }

  package {
    filename         = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domain)
}

func TestResourceFastlyFlattenFOS(t *testing.T) {
	cases := []struct {
		remote []*gofastly.FOS
		local  []any
		want   []map[string]any
	}{
		{
			remote: []*gofastly.FOS{{
				AccessKey:        new("server-generated"),
				BucketName:       new("bucket"),
				CompressionCodec: new("zstd"),
				Format:           new(LoggingFOSDefaultFormat),
				FormatVersion:    new(2),
				GzipLevel:        new(0), // API default when not set in the request.
				MessageType:      new("classic"),
				Name:             new("fos-endpoint"),
				Path:             new(""),
				Period:           new(3600),
				Region:           new("us-east"),
				TimestampFormat:  new("%Y"),
			}},
			local: []any{map[string]any{"name": "fos-endpoint", "gzip_level": -1}},
			want: []map[string]any{{
				"name":              "fos-endpoint",
				"bucket_name":       "bucket",
				"compression_codec": "zstd",
				"format":            LoggingFOSDefaultFormat,
				"format_version":    2,
				"gzip_level":        -1, // preserved from local state to avoid a perpetual diff
				"message_type":      "classic",
				"period":            3600,
				"region":            "us-east",
				"timestamp_format":  "%Y",
			}},
		},
	}

	for _, c := range cases {
		got := flattenFOS(c.remote, c.local)
		if diff := cmp.Diff(c.want, got); diff != "" {
			t.Fatalf("flattenFOS mismatch (-want +got):\n%s", diff)
		}
	}
}
