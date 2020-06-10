package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

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
	}

	for _, c := range cases {
		out := flattenBigQuery(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_bigquerylogging(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("Failed to generate key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_bigquery(name, bqName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_bq(&service, name, bqName),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_bigquerylogging_env(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	bqName := fmt.Sprintf("bq %s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("Failed to generate key: %s", err)
	}

	// set env Vars to something we expect
	resetEnv := setBQEnv("someEnv", secretKey, t)
	defer resetEnv()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_bigquery_env(name, bqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_bq(&service, name, bqName),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_bq(service *gofastly.ServiceDetail, name, bqName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		bqList, err := conn.ListBigQueries(&gofastly.ListBigQueriesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
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

func testAccServiceV1Config_bigquery(name, gcsName, secretKey string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  bigquerylogging {
    name       = "%s"
    email      = "email@example.com"
    secret_key = %q
    project_id = "example-gcp-project"
    dataset    = "example_bq_dataset"
    table      = "example_bq_table"
  }

  force_destroy = true
}`, name, domainName, backendName, gcsName, secretKey)
}

func testAccServiceV1Config_bigquery_env(name, gcsName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  bigquerylogging {
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
