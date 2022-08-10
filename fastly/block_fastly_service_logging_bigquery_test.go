package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceVCL_bigquerylogging(t *testing.T) {
	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("Failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfig_bigquery(name, bqName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributes_bq(&service, name, bqName),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_bigquerylogging_compute(t *testing.T) {
	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("Failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfig_bigquery_compute(name, bqName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLAttributes_bq(&service, name, bqName),
				),
			},
		},
	})
}

func TestBigqueryloggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingBigQuery(serviceAttributes)
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	v.Register(resource)
	loggingResource := resource.Schema["logging_bigquery"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["email"].Sensitive {
		t.Fatalf("Expected email to be marked as a Sensitive value")
	}

	if !loggingResourceSchema["secret_key"].Sensitive {
		t.Fatalf("Expected secret_key to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	email := "tf-test@fastly.com"
	secretKey, _ := generateKey()
	resetEnv := setBQEnv(email, secretKey, t)
	defer resetEnv()

	result1, err1 := loggingResourceSchema["email"].DefaultFunc()
	if err1 != nil {
		t.Fatalf("Unexpected err %#v when calling email DefaultFunc", err1)
	}
	if result1 != email {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", email, result1)
	}

	result2, err2 := loggingResourceSchema["secret_key"].DefaultFunc()
	if err2 != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err2)
	}
	if result2 != secretKey {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", secretKey, result2)
	}
}

func testAccCheckFastlyServiceVCLAttributes_bq(service *gofastly.ServiceDetail, name, bqName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		bqList, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up BigQuery records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(bqList) != 1 {
			return fmt.Errorf("BigQuery logging endpoint missing, expected: 1, got: %d", len(bqList))
		}

		if bqList[0].Name != bqName {
			return fmt.Errorf("BigQuery logging endpoint name mismatch, expected: %s, got: %#v", bqName, bqList[0].Name)
		}

		return nil
	}
}

func testAccServiceVCLConfig_bigquery(name, gcsName, secretKey string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  logging_bigquery {
    name       = "%s"
    email      = "email@example.com"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
	format     = "%%h %%l %%u %%t %%r %%>s"
	placement  = "waf_debug"
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey)
}

func testAccServiceVCLConfig_bigquery_compute(name, gcsName, secretKey string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  logging_bigquery {
    name       = "%s"
    email      = "email@example.com"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey)
}

func testAccServiceVCLConfig_bigquery_env(name, gcsName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  logging_bigquery {
    name       = "%s"
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName)
}

func setBQEnv(email, secretKey string, t *testing.T) func() {
	e := getBQEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_BQ_EMAIL", email); err != nil {
		t.Fatalf("Error setting env var FASTLY_BQ_EMAIL: %s", err)
	}
	if err := os.Setenv("FASTLY_BQ_SECRET_KEY", secretKey); err != nil {
		t.Fatalf("Error setting env var FASTLY_BQ_SECRET_KEY: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_BQ_EMAIL", e.Key); err != nil {
			t.Fatalf("Error resetting env var FASTLY_BQ_EMAIL: %s", err)
		}
		if err := os.Setenv("FASTLY_BQ_SECRET_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var FASTLY_BQ_SECRET_KEY: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentBQEnv struct {
	Key, Secret string
}

func getBQEnv() *currentBQEnv {
	// Grab any existing Fastly BigQuery keys and preserve, in the off chance
	// they're actually set in the enviornment
	return &currentBQEnv{
		Key:    os.Getenv("FASTLY_BQ_SECRET_KEY"),
		Secret: os.Getenv("FASTLY_BQ_SECRET_KEY"),
	}
}

// TestResourceFastlyFlattenBigQuery tests the flattenBigQuery function
func TestResourceFastlyFlattenBigQuery(t *testing.T) {
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("Failed to generate key: %s", err)
	}

	cases := []struct {
		remote []*gofastly.BigQuery
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.BigQuery{
				{
					Name:      "bigquery-example",
					User:      "email@example.com",
					ProjectID: "example-gcp-project",
					Dataset:   "example_bq_dataset",
					Table:     "example_bq_table",
					SecretKey: secretKey,
				},
			},
			local: []map[string]interface{}{
				{
					"name":       "bigquery-example",
					"email":      "email@example.com",
					"project_id": "example-gcp-project",
					"dataset":    "example_bq_dataset",
					"table":      "example_bq_table",
					"secret_key": secretKey,
				},
			},
		},
		{
			remote: []*gofastly.BigQuery{
				{
					Name:              "bigquery-example",
					User:              "email@example.com",
					ProjectID:         "example-gcp-project",
					Dataset:           "example_bq_dataset",
					Table:             "example_bq_table",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					Placement:         "waf_debug",
					ResponseCondition: "error_response",
					SecretKey:         secretKey,
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "bigquery-example",
					"email":              "email@example.com",
					"project_id":         "example-gcp-project",
					"dataset":            "example_bq_dataset",
					"table":              "example_bq_table",
					"secret_key":         secretKey,
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"placement":          "waf_debug",
					"response_condition": "error_response",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBigQuery(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}
