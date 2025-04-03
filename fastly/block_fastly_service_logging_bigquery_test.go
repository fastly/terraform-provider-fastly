package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceVCL_bigquerylogging(t *testing.T) {
	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))
	email := "email@example.com"
	emailUpdate := "update@example.com"

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBigQuery(name, bqName, secretKey, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBQ(&service, name, bqName, email),
				),
			},
			{
				Config: testAccServiceVCLConfigBigQuery(name, bqName, secretKey, emailUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBQ(&service, name, bqName, emailUpdate),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_bigquerylogging_compute(t *testing.T) {
	var service gofastly.ServiceDetail

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))
	email := "email@example.com"
	emailUpdate := "update@example.com"

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBigQueryCompute(name, bqName, secretKey, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBQ(&service, name, bqName, email),
				),
			},
			{
				Config: testAccServiceVCLConfigBigQueryCompute(name, bqName, secretKey, emailUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBQ(&service, name, bqName, emailUpdate),
				),
			},
		},
	})
}

func TestBigqueryloggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingBigQuery(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_bigquery"]
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

func testAccCheckFastlyServiceVCLAttributesBQ(service *gofastly.ServiceDetail, name, bqName, email string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		bqList, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up BigQuery records for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(bqList) != 1 {
			return fmt.Errorf("bigQuery logging endpoint missing, expected: 1, got: %d", len(bqList))
		}

		if gofastly.ToValue(bqList[0].Name) != bqName {
			return fmt.Errorf("bigQuery logging endpoint name mismatch, expected: %s, got: %#v", bqName, gofastly.ToValue(bqList[0].Name))
		}

		if gofastly.ToValue(bqList[0].User) != email {
			return fmt.Errorf("bigQuery logging endpoint user/email mismatch, expected: %s, got: %#v", email, gofastly.ToValue(bqList[0].User))
		}

		return nil
	}
}

func testAccServiceVCLConfigBigQuery(name, gcsName, secretKey, email string) string {
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
    secret_key = trimspace(%q)
    email      = "%s"
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
    format     = "%%h %%l %%u %%t %%r %%>s"
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey, email)
}

func testAccServiceVCLConfigBigQueryCompute(name, gcsName, secretKey, email string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
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
    address = "%s"
    name    = "tf -test backend"
  }

  logging_bigquery {
    name       = "%s"
    secret_key = trimspace(%q)
    email      = "%s"
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey, email)
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
	// they're actually set in the environment
	return &currentBQEnv{
		Key:    os.Getenv("FASTLY_BQ_SECRET_KEY"),
		Secret: os.Getenv("FASTLY_BQ_SECRET_KEY"),
	}
}

// TestResourceFastlyFlattenBigQuery tests the flattenBigQuery function
func TestResourceFastlyFlattenBigQuery(t *testing.T) {
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	cases := []struct {
		remote []*gofastly.BigQuery
		local  []map[string]any
	}{
		{
			remote: []*gofastly.BigQuery{
				{
					Name:      gofastly.ToPointer("bigquery-example"),
					User:      gofastly.ToPointer("email@example.com"),
					ProjectID: gofastly.ToPointer("example-gcp-project"),
					Dataset:   gofastly.ToPointer("example_bq_dataset"),
					Table:     gofastly.ToPointer("example_bq_table"),
					SecretKey: gofastly.ToPointer(secretKey),
				},
			},
			local: []map[string]any{
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
					Name:              gofastly.ToPointer("bigquery-example"),
					User:              gofastly.ToPointer("email@example.com"),
					ProjectID:         gofastly.ToPointer("example-gcp-project"),
					Dataset:           gofastly.ToPointer("example_bq_dataset"),
					Table:             gofastly.ToPointer("example_bq_table"),
					Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("error_response"),
					SecretKey:         gofastly.ToPointer(secretKey),
				},
			},
			local: []map[string]any{
				{
					"name":               "bigquery-example",
					"email":              "email@example.com",
					"project_id":         "example-gcp-project",
					"dataset":            "example_bq_dataset",
					"table":              "example_bq_table",
					"secret_key":         secretKey,
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"placement":          "none",
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
