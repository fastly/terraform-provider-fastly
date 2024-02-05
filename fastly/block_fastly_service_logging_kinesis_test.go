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

const testKinesisIAMRole = "arn:aws:iam::123456789012:role/KinesisAccess"

func TestResourceFastlyFlattenKinesis(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Kinesis
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("kinesis-endpoint"),
					StreamName:        gofastly.ToPointer("stream-name"),
					Region:            gofastly.ToPointer("us-east-1"),
					AccessKey:         gofastly.ToPointer("whywouldyoucheckthis"),
					SecretKey:         gofastly.ToPointer("thisisthesecretthatneedstobe40characters"),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					FormatVersion:     gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"access_key":         "whywouldyoucheckthis",
					"secret_key":         "thisisthesecretthatneedstobe40characters",
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
					"placement":          "none",
					"response_condition": "always",
					"format_version":     2,
				},
			},
		},
		{
			remote: []*gofastly.Kinesis{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("kinesis-endpoint"),
					StreamName:        gofastly.ToPointer("stream-name"),
					Region:            gofastly.ToPointer("us-east-1"),
					IAMRole:           gofastly.ToPointer(testKinesisIAMRole),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("always"),
					FormatVersion:     gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":               "kinesis-endpoint",
					"topic":              "stream-name",
					"region":             "us-east-1",
					"iam_role":           testKinesisIAMRole,
					"format":             "%h %l %u %t \"%r\" %>s %b %T",
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

func TestAccFastlyServiceVCL_logging_kinesis_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		AccessKey:         gofastly.ToPointer("whywouldyoucheckthis"),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		IAMRole:           gofastly.ToPointer(""),
		Name:              gofastly.ToPointer("kinesis-endpoint"),
		Region:            gofastly.ToPointer("us-east-1"),
		ResponseCondition: gofastly.ToPointer(""),
		SecretKey:         gofastly.ToPointer("thisisthesecretthatneedstobe40characters"),
		ServiceVersion:    gofastly.ToPointer(1),
		StreamName:        gofastly.ToPointer("stream-name"),
	}

	log1AfterUpdate := gofastly.Kinesis{
		AccessKey:         gofastly.ToPointer(""),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		IAMRole:           gofastly.ToPointer(testKinesisIAMRole),
		Name:              gofastly.ToPointer("kinesis-endpoint"),
		Region:            gofastly.ToPointer("us-east-1"),
		ResponseCondition: gofastly.ToPointer(""),
		SecretKey:         gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		StreamName:        gofastly.ToPointer("new-stream-name"),
	}

	log2 := gofastly.Kinesis{
		AccessKey:         gofastly.ToPointer(""),
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		IAMRole:           gofastly.ToPointer(testKinesisIAMRole),
		Name:              gofastly.ToPointer("another-kinesis-endpoint"),
		Region:            gofastly.ToPointer("us-east-1"),
		ResponseCondition: gofastly.ToPointer(""),
		SecretKey:         gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		StreamName:        gofastly.ToPointer("another-stream-name"),
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

func TestAccFastlyServiceVCL_logging_kinesis_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Kinesis{
		AccessKey:      gofastly.ToPointer("whywouldyoucheckthis"),
		IAMRole:        gofastly.ToPointer(""),
		Name:           gofastly.ToPointer("kinesis-endpoint"),
		Region:         gofastly.ToPointer("us-east-1"),
		SecretKey:      gofastly.ToPointer("thisisthesecretthatneedstobe40characters"),
		ServiceVersion: gofastly.ToPointer(1),
		StreamName:     gofastly.ToPointer("stream-name"),
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
		ksl, err := conn.ListKinesis(&gofastly.ListKinesisInput{
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
    format      = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLKinesisConfigUpdate(name, domain string) string {
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
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}
