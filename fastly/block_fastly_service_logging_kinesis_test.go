package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testKinesisIAMRole = "arn:aws:iam::123456789012:role/KinesisAccess"

func TestResourceFastlyFlattenKinesis(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kinesis
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    1,
					Name:              "kinesis-endpoint",
					StreamName:        "stream-name",
					Region:            "us-east-1",
					AccessKey:         "whywouldyoucheckthis",
					SecretKey:         "thisisthesecretthatneedstobe40characters",
					Format:            "%h %l %u %t \"%r\" %>s %b %T",
					Placement:         "none",
					ResponseCondition: "always",
					FormatVersion:     2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"access_key":         "whywouldyoucheckthis",
					"secret_key":         "thisisthesecretthatneedstobe40characters",
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
					"placement":          "none",
					"response_condition": "always",
					"format_version":     uint(2),
				},
			},
		},
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    1,
					Name:              "kinesis-endpoint",
					StreamName:        "stream-name",
					Region:            "us-east-1",
					IAMRole:           testKinesisIAMRole,
					Format:            "%h %l %u %t \"%r\" %>s %b %T",
					Placement:         "none",
					ResponseCondition: "always",
					FormatVersion:     2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"iam_role":           testKinesisIAMRole,
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
					"placement":          "none",
					"response_condition": "always",
					"format_version":     uint(2),
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

func TestAccFastlyServiceVCL_logging_kinesis_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		ServiceVersion: 1,
		Name:           "kinesis-endpoint",
		StreamName:     "stream-name",
		Region:         "us-east-1",
		AccessKey:      "whywouldyoucheckthis",
		SecretKey:      "thisisthesecretthatneedstobe40characters",
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	log1_after_update := gofastly.Kinesis{
		ServiceVersion: 1,
		Name:           "kinesis-endpoint",
		StreamName:     "new-stream-name",
		Region:         "us-east-1",
		IAMRole:        testKinesisIAMRole,
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.Kinesis{
		ServiceVersion: 1,
		Name:           "another-kinesis-endpoint",
		StreamName:     "another-stream-name",
		Region:         "us-east-1",
		IAMRole:        testKinesisIAMRole,
		FormatVersion:  2,
		Format:         "%h %l %u %t \"%r\" %>s %b",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLKinesisConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_kinesis.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLKinesisConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_kinesis.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_kinesis_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		ServiceVersion: 1,
		Name:           "kinesis-endpoint",
		StreamName:     "stream-name",
		Region:         "us-east-1",
		AccessKey:      "whywouldyoucheckthis",
		SecretKey:      "thisisthesecretthatneedstobe40characters",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLKinesisComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLKinesisAttributes(&service, []*gofastly.Kinesis{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_kinesis.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLKinesisAttributes(service *gofastly.ServiceDetail, Kinesis []*gofastly.Kinesis, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		KinesisList, err := conn.ListKinesis(&gofastly.ListKinesisInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Kinesis Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(KinesisList) != len(Kinesis) {
			return fmt.Errorf("Kinesis List count mismatch, expected (%d), got (%d)", len(Kinesis), len(KinesisList))
		}

		log.Printf("[DEBUG] KinesisList = %#v\n", KinesisList)

		for _, e := range Kinesis {
			for _, el := range KinesisList {
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
						return fmt.Errorf("Bad match Kinesis logging match: %s", diff)
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
    format      = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKinesisConfig_update(name, domain string) string {
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
    format      = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
  }

  logging_kinesis {
    name        = "another-kinesis-endpoint"
    topic       = "another-stream-name"
    region      = "us-east-1"
    iam_role    = "%s"
    format      = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain, testKinesisIAMRole, testKinesisIAMRole)
}

func testAccServiceVCLKinesisComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}
`, name, domain)
}
