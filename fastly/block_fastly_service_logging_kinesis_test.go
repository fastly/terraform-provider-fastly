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

const testKinesisIAMRole = "arn:aws:iam::123456789012:role/KinesisAccess"

func TestAccFastlyServiceLoggingKinesis_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		AccessKey:         new("whywouldyoucheckthis"),
		Format:            new(LoggingKinesisDefaultFormat),
		FormatVersion:     new(2),
		IAMRole:           new(""),
		Name:              new("kinesis-endpoint"),
		Region:            new("us-east-1"),
		ResponseCondition: new(""),
		SecretKey:         new("thisisthesecretthatneedstobe40characters"),
		ServiceVersion:    new(1),
		StreamName:        new("stream-name"),
		ProcessingRegion:  new("us"),
	}

	log1AfterUpdate := gofastly.Kinesis{
		AccessKey:         new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		IAMRole:           new(testKinesisIAMRole),
		Name:              new("kinesis-endpoint"),
		Region:            new("us-east-1"),
		ResponseCondition: new(""),
		SecretKey:         new(""),
		ServiceVersion:    new(1),
		StreamName:        new("new-stream-name"),
		ProcessingRegion:  new("none"),
	}

	log2 := gofastly.Kinesis{
		AccessKey:         new(""),
		Format:            new(LoggingFormatUpdate),
		FormatVersion:     new(2),
		IAMRole:           new(testKinesisIAMRole),
		Name:              new("another-kinesis-endpoint"),
		Region:            new("us-east-1"),
		ResponseCondition: new(""),
		SecretKey:         new(""),
		ServiceVersion:    new(1),
		StreamName:        new("another-stream-name"),
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
				Config: testAccServiceVCLKinesisConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kinesis.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLKinesisConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_kinesis.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingKinesis_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		AccessKey:        new("whywouldyoucheckthis"),
		IAMRole:          new(""),
		Name:             new("kinesis-endpoint"),
		Region:           new("us-east-1"),
		SecretKey:        new("thisisthesecretthatneedstobe40characters"),
		ServiceVersion:   new(1),
		StreamName:       new("stream-name"),
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
				Config: testAccServiceVCLKinesisComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_kinesis.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLKinesisAttributes(service *gofastly.ServiceDetail, ks []*gofastly.Kinesis, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		ksl, err := conn.ListKinesis(context.TODO(), &gofastly.ListKinesisInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Kinesis Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(ksl) != len(ks) {
			return fmt.Errorf("kinesis List count mismatch, expected (%d), got (%d)", len(ks), len(ksl))
		}

		log.Printf("[DEBUG] KinesisList = %#v\n", ksl)

		for _, e := range ks {
			for _, el := range ksl {
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
						return fmt.Errorf("bad match Kinesis logging match: %s", diff)
					}
				}
			}
		}

		return nil
	}
}

func testAccServiceVCLKinesisConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-kinesis-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_kinesis {
    name        = "kinesis-endpoint"
    topic       = "stream-name"
    region      = "us-east-1"
    access_key  = "whywouldyoucheckthis"
    secret_key  = "thisisthesecretthatneedstobe40characters"
    processing_region = "us"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKinesisConfigUpdate(name, domain string) string {
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-kinesis-logging"
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

  logging_kinesis {
    name        = "kinesis-endpoint"
    topic       = "new-stream-name"
    region      = "us-east-1"
    iam_role    = "%s"
    format      = %q
  }

  logging_kinesis {
    name        = "another-kinesis-endpoint"
    topic       = "another-stream-name"
    region      = "us-east-1"
    iam_role    = "%s"
    format      = %q
  }

  force_destroy = true
}
`, name, domain, testKinesisIAMRole, format, testKinesisIAMRole, format)
}

func testAccServiceVCLKinesisComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-kinesis-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_kinesis {
    name        = "kinesis-endpoint"
    topic       = "stream-name"
    region      = "us-east-1"
    access_key  = "whywouldyoucheckthis"
    secret_key  = "thisisthesecretthatneedstobe40characters"
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

func TestResourceFastlyFlattenKinesis(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kinesis
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    new(1),
					Name:              new("kinesis-endpoint"),
					StreamName:        new("stream-name"),
					Region:            new("us-east-1"),
					AccessKey:         new("whywouldyoucheckthis"),
					SecretKey:         new("thisisthesecretthatneedstobe40characters"),
					Format:            new(LoggingKinesisDefaultFormat),
					Placement:         new("none"),
					ResponseCondition: new("always"),
					FormatVersion:     new(2),
					ProcessingRegion:  new("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"access_key":         "whywouldyoucheckthis",
					"secret_key":         "thisisthesecretthatneedstobe40characters",
					"format":             LoggingKinesisDefaultFormat,
					"placement":          "none",
					"response_condition": "always",
					"format_version":     2,
					"processing_region":  "eu",
				},
			},
		},
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    new(1),
					Name:              new("kinesis-endpoint"),
					StreamName:        new("stream-name"),
					Region:            new("us-east-1"),
					IAMRole:           new(testKinesisIAMRole),
					Format:            new(LoggingFormatUpdate),
					Placement:         new("none"),
					ResponseCondition: new("always"),
					FormatVersion:     new(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"iam_role":           testKinesisIAMRole,
					"format":             LoggingFormatUpdate,
					"placement":          "none",
					"response_condition": "always",
					"format_version":     2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenKinesis(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
