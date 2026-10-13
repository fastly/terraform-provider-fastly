package fastly

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func TestAccFastlyServiceLoggingBigQuery_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	email := "email@example.com"
	emailUpdate := "update@example.com"

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	bigQueryLogOne := gofastly.BigQuery{
		Dataset:           new("example_bq_dataset"),
		User:              new(email),
		Format:            new(LoggingBigQueryDefaultFormat),
		Name:              new("test-bigquery-1"),
		Placement:         new("none"),
		ProjectID:         new("example-gcp-project"),
		ResponseCondition: new("error_response_5XX"),
		SecretKey:         new(secretKey),
		Table:             new("example_bq_table"),
		Template:          new("_1"),
		ProcessingRegion:  new("us"),
	}

	bigQueryLogOneUpdated := gofastly.BigQuery{
		Dataset:           new("example_bq_dataset"),
		User:              new(emailUpdate),
		Format:            new(LoggingFormatUpdate),
		Name:              new("test-bigquery-1"),
		Placement:         new("none"),
		ProjectID:         new("example-gcp-project"),
		ResponseCondition: new("error_response_5XX"),
		SecretKey:         new(secretKey),
		Table:             new("example_bq_table"),
		Template:          new("_1_updated"),
		ProcessingRegion:  new("none"),
	}

	bigQueryLogTwo := gofastly.BigQuery{
		Dataset:           new("example_bq_dataset_2"),
		User:              new(emailUpdate),
		Format:            new(LoggingFormatUpdate),
		Name:              new("test-bigquery-2"),
		Placement:         new("none"),
		ProjectID:         new("example-gcp-project-2"),
		ResponseCondition: new("ok_response_2XX"),
		SecretKey:         new(secretKey),
		Table:             new("example_bq_table_2"),
		Template:          new("_2"),
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
				Config: testAccServiceVCLBigQueryLoggingConfigComplete(serviceName, email, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBigQueryLoggingAttributes(&service, []*gofastly.BigQuery{&bigQueryLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_bigquery.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLBigQueryLoggingConfigUpdate(serviceName, emailUpdate, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBigQueryLoggingAttributes(&service, []*gofastly.BigQuery{&bigQueryLogOneUpdated, &bigQueryLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_bigquery.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingBigQuery_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	email := "email@example.com"

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	bigQueryLogOne := gofastly.BigQuery{
		Dataset:          new("example_bq_dataset"),
		User:             new(email),
		Name:             new("test-bigquery-1"),
		ProjectID:        new("example-gcp-project"),
		SecretKey:        new(secretKey),
		Table:            new("example_bq_table"),
		Template:         new("_1"),
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
				Config: testAccServiceVCLBigQueryLoggingConfigCompleteCompute(serviceName, email, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLBigQueryLoggingAttributes(&service, []*gofastly.BigQuery{&bigQueryLogOne}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_bigquery.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingBigQuery_vcl_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	email := "email@example.com"

	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	bigQueryLog := gofastly.BigQuery{
		Dataset:           new("example_bq_dataset"),
		User:              new(email),
		Format:            new(LoggingBigQueryDefaultFormat),
		Name:              new("test-bigquery"),
		ProjectID:         new("example-gcp-project"),
		ResponseCondition: new(""),
		SecretKey:         new(secretKey),
		Table:             new("example_bq_table"),
		Template:          new(""),
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
				Config: testAccServiceVCLBigQueryLoggingConfigDefault(serviceName, email, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBigQueryLoggingAttributes(&service, []*gofastly.BigQuery{&bigQueryLog}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_bigquery.#", "1"),
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

	emailResult, err := loggingResourceSchema["email"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling email DefaultFunc", err)
	}
	if emailResult != email {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", email, emailResult)
	}

	secretkeyResult, err := loggingResourceSchema["secret_key"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling secret_key DefaultFunc", err)
	}
	if secretkeyResult != secretKey {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", secretKey, secretkeyResult)
	}

	formatSchema, ok := loggingResourceSchema["format"]
	if !ok {
		t.Fatalf("Expected format field to exist in schema")
	}
	if formatSchema.Default != LoggingBigQueryDefaultFormat {
		t.Fatalf("Error matching format default:\nexpected: %#v\ngot: %#v", LoggingBigQueryDefaultFormat, formatSchema.Default)
	}
}

func testAccCheckFastlyServiceVCLBigQueryLoggingAttributes(service *gofastly.ServiceDetail, localBigQueryList []*gofastly.BigQuery, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		remoteBigQueryList, err := conn.ListBigQueries(context.TODO(), &gofastly.ListBigQueriesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up BigQuery Logging for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(remoteBigQueryList) != len(localBigQueryList) {
			return fmt.Errorf("bigQuery List count mismatch, expected (%d), got (%d)", len(localBigQueryList), len(remoteBigQueryList))
		}

		var found int
		for _, lbq := range localBigQueryList {
			for _, rbq := range remoteBigQueryList {
				if gofastly.ToValue(lbq.Name) == gofastly.ToValue(rbq.Name) {
					// we don't know these things ahead of time, so populate them now
					lbq.ServiceID = service.ServiceID
					lbq.ServiceVersion = service.ActiveVersion.Number

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						lbq.Format = rbq.Format
						lbq.ResponseCondition = rbq.ResponseCondition
						lbq.Placement = rbq.Placement
					}

					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					rbq.CreatedAt = nil
					rbq.UpdatedAt = nil
					// BigQuery API may return FormatVersion but we don't support it in schema
					rbq.FormatVersion = nil
					if diff := cmp.Diff(lbq, rbq); diff != "" {
						return fmt.Errorf("bad match BigQuery logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(localBigQueryList) {
			return fmt.Errorf("error matching BigQuery Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLBigQueryLoggingConfigComplete(serviceName, email, secretKey string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "tf-test-backend"
  }

  condition {
    name = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority = 10
    type = "RESPONSE"
  }

  logging_bigquery {
    name = "test-bigquery-1"
    email = "%s"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset = "example_bq_dataset"
    table = "example_bq_table"
    template = "_1"
    placement = "none"
    response_condition = "error_response_5XX"
    processing_region = "us"
  }

  force_destroy = true
}`, serviceName, domainName, email, secretKey)
}

func testAccServiceVCLBigQueryLoggingConfigCompleteCompute(serviceName, email, secretKey string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "tf-test-backend"
  }

  logging_bigquery {
    name = "test-bigquery-1"
    email = "%s"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset = "example_bq_dataset"
    table = "example_bq_table"
    template = "_1"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, serviceName, domainName, email, secretKey)
}

func testAccServiceVCLBigQueryLoggingConfigUpdate(serviceName, email, secretKey string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	format := LoggingFormatUpdate
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "tf-test-backend"
  }

  condition {
    name = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority = 10
    type = "RESPONSE"
  }

  condition {
    name = "ok_response_2XX"
    statement = "resp.status >= 200 && resp.status < 300"
    priority = 10
    type = "RESPONSE"
  }

  logging_bigquery {
    name = "test-bigquery-1"
    email = "%s"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset = "example_bq_dataset"
    table = "example_bq_table"
    template = "_1_updated"
    format = %q
    placement = "none"
    response_condition = "error_response_5XX"
  }

  logging_bigquery {
    name = "test-bigquery-2"
    email = "%s"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project-2"
    dataset = "example_bq_dataset_2"
    table = "example_bq_table_2"
    template = "_2"
    format = %q
    placement = "none"
    response_condition = "ok_response_2XX"
  }

  force_destroy = true
}`, serviceName, domainName, email, secretKey, format, email, secretKey, format)
}

func testAccServiceVCLBigQueryLoggingConfigDefault(serviceName, email, secretKey string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "tf-test-backend"
  }

  logging_bigquery {
    name = "test-bigquery"
    email = "%s"
    secret_key = trimspace(%q)
    project_id = "example-gcp-project"
    dataset = "example_bq_dataset"
    table = "example_bq_table"
  }

  force_destroy = true
}`, serviceName, domainName, email, secretKey)
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
		if err := os.Setenv("FASTLY_BQ_EMAIL", e.Email); err != nil {
			t.Fatalf("Error resetting env var FASTLY_BQ_EMAIL: %s", err)
		}
		if err := os.Setenv("FASTLY_BQ_SECRET_KEY", e.SecretKey); err != nil {
			t.Fatalf("Error resetting env var FASTLY_BQ_SECRET_KEY: %s", err)
		}
	}
}

// struct to preserve the current environment.
type currentBQEnv struct {
	Email, SecretKey string
}

func getBQEnv() *currentBQEnv {
	// Grab any existing Fastly BigQuery keys and preserve, in the off chance
	// they're actually set in the environment
	return &currentBQEnv{
		Email:     os.Getenv("FASTLY_BQ_EMAIL"),
		SecretKey: os.Getenv("FASTLY_BQ_SECRET_KEY"),
	}
}

// TestResourceFastlyFlattenBigQuery tests the flattenBigQuery function.
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
					Name:      new("bigquery-example"),
					User:      new("email@example.com"),
					ProjectID: new("example-gcp-project"),
					Dataset:   new("example_bq_dataset"),
					Table:     new("example_bq_table"),
					Format:    new(LoggingBigQueryDefaultFormat),
					SecretKey: new(secretKey),
					Template:  new("_1"),
				},
			},
			local: []map[string]any{
				{
					"name":       "bigquery-example",
					"email":      "email@example.com",
					"project_id": "example-gcp-project",
					"dataset":    "example_bq_dataset",
					"table":      "example_bq_table",
					"format":     LoggingBigQueryDefaultFormat,
					"secret_key": secretKey,
					"template":   "_1",
				},
			},
		},
		{
			remote: []*gofastly.BigQuery{
				{
					Name:              new("bigquery-example"),
					User:              new("email@example.com"),
					ProjectID:         new("example-gcp-project"),
					Dataset:           new("example_bq_dataset"),
					Table:             new("example_bq_table"),
					Format:            new(LoggingFormatUpdate),
					Placement:         new("none"),
					ResponseCondition: new("error_response"),
					SecretKey:         new(secretKey),
					Template:          new("_updated"),
					ProcessingRegion:  new("eu"),
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
					"format":             LoggingFormatUpdate,
					"placement":          "none",
					"response_condition": "error_response",
					"template":           "_updated",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBigQuery(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
