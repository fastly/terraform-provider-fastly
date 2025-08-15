package fastly

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestAccFastlyServiceVCL_gcslogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	gcsLogOne := gofastly.GCS{
		AccountName:       gofastly.ToPointer("service-account"),
		Bucket:            gofastly.ToPointer("bucketname"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer(LoggingGCSDefaultFormat),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("test-gcs-1"),
		Path:              gofastly.ToPointer("/5XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		ProjectID:         gofastly.ToPointer("project-id"),
		ResponseCondition: gofastly.ToPointer("error_response_5XX"),
		SecretKey:         gofastly.ToPointer(secretKey),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("email@example.com"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	gcsLogOneUpdated := gofastly.GCS{
		AccountName:       gofastly.ToPointer("service-account"),
		Bucket:            gofastly.ToPointer("bucketname"),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(1),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("test-gcs-1"),
		Path:              gofastly.ToPointer("/5XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		ProjectID:         gofastly.ToPointer("project-id"),
		ResponseCondition: gofastly.ToPointer("error_response_5XX"),
		SecretKey:         gofastly.ToPointer(secretKey),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("email@example.com"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	gcsLogTwo := gofastly.GCS{
		AccountName:       gofastly.ToPointer("service-account"),
		Bucket:            gofastly.ToPointer("bucketname"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0),
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("test-gcs-2"),
		Path:              gofastly.ToPointer("/2XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		ProjectID:         gofastly.ToPointer("project-id"),
		ResponseCondition: gofastly.ToPointer("ok_response_2XX"),
		SecretKey:         gofastly.ToPointer(secretKey),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:              gofastly.ToPointer("email@example.com"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLGCSLoggingConfigComplete(serviceName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGCSLoggingAttributes(&service, []*gofastly.GCS{&gcsLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_gcs.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLGCSLoggingConfigUpdate(serviceName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGCSLoggingAttributes(&service, []*gofastly.GCS{&gcsLogOneUpdated, &gcsLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_gcs.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_gcslogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	gcsLogOne := gofastly.GCS{
		AccountName:      gofastly.ToPointer("service-account"),
		Bucket:           gofastly.ToPointer("bucketname"),
		CompressionCodec: gofastly.ToPointer("zstd"),
		GzipLevel:        gofastly.ToPointer(0),
		MessageType:      gofastly.ToPointer("classic"),
		Name:             gofastly.ToPointer("test-gcs-1"),
		Path:             gofastly.ToPointer("/5XX/"),
		Period:           gofastly.ToPointer(12),
		ProjectID:        gofastly.ToPointer("project-id"),
		SecretKey:        gofastly.ToPointer(secretKey),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		User:             gofastly.ToPointer("email@example.com"),
		ProcessingRegion: gofastly.ToPointer("us"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLGCSLoggingConfigCompleteCompute(serviceName, secretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLGCSLoggingAttributes(&service, []*gofastly.GCS{&gcsLogOne}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_gcs.#", "1"),
				),
			},
		},
	})
}

func TestGcsloggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingGCS(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_gcs"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["secret_key"].Sensitive {
		t.Fatalf("Expected secret_key to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	email := "tf-test@fastly.com"
	secretKey, _ := generateKey()
	resetEnv := setGcsEnv(email, secretKey, t)
	defer resetEnv()

	result1, err1 := loggingResourceSchema["user"].DefaultFunc()
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

func testAccCheckFastlyServiceVCLGCSLoggingAttributes(service *gofastly.ServiceDetail, localGCSList []*gofastly.GCS, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		remoteGCSList, err := conn.ListGCSs(context.TODO(), &gofastly.ListGCSsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up GCS Logging for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(remoteGCSList) != len(localGCSList) {
			return fmt.Errorf("GCS List count mismatch, expected (%d), got (%d)", len(localGCSList), len(remoteGCSList))
		}

		var found int
		for _, lgcs := range localGCSList {
			for _, rgcs := range remoteGCSList {
				if gofastly.ToValue(lgcs.Name) == gofastly.ToValue(rgcs.Name) {
					// we don't know these things ahead of time, so populate them now
					lgcs.ServiceID = service.ServiceID
					lgcs.ServiceVersion = service.ActiveVersion.Number

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						lgcs.FormatVersion = rgcs.FormatVersion
						lgcs.Format = rgcs.Format
						lgcs.ResponseCondition = rgcs.ResponseCondition
						lgcs.Placement = rgcs.Placement
					}

					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					rgcs.CreatedAt = nil
					rgcs.UpdatedAt = nil
					if diff := cmp.Diff(lgcs, rgcs); diff != "" {
						return fmt.Errorf("bad match GCS logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(localGCSList) {
			return fmt.Errorf("error matching GCS Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLGCSLoggingConfigComplete(serviceName, secretKey string) string {
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

  logging_gcs {
    name = "test-gcs-1"
    user = "email@example.com"
    bucket_name = "bucketname"
    account_name = "service-account"
    project_id = "project-id"
    secret_key = %q
    path = "/5XX/"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    format_version = 2
    placement = "none"
    response_condition = "error_response_5XX"
    compression_codec = "zstd"
    processing_region = "us"
  }

  force_destroy = true
}`, serviceName, domainName, secretKey)
}

func testAccServiceVCLGCSLoggingConfigCompleteCompute(serviceName, secretKey string) string {
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

  logging_gcs {
    account_name = "service-account"
    bucket_name = "bucketname"
    compression_codec = "zstd"
    name = "test-gcs-1"
    path = "/5XX/"
    period = 12
    project_id = "project-id"
    secret_key = %q
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    user = "email@example.com"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, serviceName, domainName, secretKey)
}

func testAccServiceVCLGCSLoggingConfigUpdate(serviceName, secretKey string) string {
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

  logging_gcs {
    name = "test-gcs-1"
    user = "email@example.com"
    bucket_name = "bucketname"
    account_name = "service-account"
    project_id = "project-id"
    secret_key = %q
    path = "/5XX/"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    format = %q
    gzip_level = 1
    format_version = 2
    placement = "none"
    response_condition = "error_response_5XX"
    processing_region = "none"
  }

  logging_gcs {
    name = "test-gcs-2"
    user = "email@example.com"
    bucket_name = "bucketname"
    account_name = "service-account"
    project_id = "project-id"
    secret_key = %q
    path = "/2XX/"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    format = %q
    format_version = 2
    placement = "none"
    response_condition = "ok_response_2XX"
    compression_codec = "zstd"
    processing_region = "none"
  }

  force_destroy = true
}`, serviceName, domainName, secretKey, format, secretKey, format)
}

func setGcsEnv(email, secretKey string, t *testing.T) func() {
	e := getGcsEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_GCS_EMAIL", email); err != nil {
		t.Fatalf("Error setting env var FASTLY_GCS_EMAIL: %s", err)
	}
	if err := os.Setenv("FASTLY_GCS_SECRET_KEY", secretKey); err != nil {
		t.Fatalf("Error setting env var FASTLY_GCS_SECRET_KEY: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_GCS_EMAIL", e.Key); err != nil {
			t.Fatalf("Error resetting env var FASTLY_GCS_EMAIL: %s", err)
		}
		if err := os.Setenv("FASTLY_GCS_SECRET_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var FASTLY_GCS_SECRET_KEY: %s", err)
		}
	}
}

// struct to preserve the current environment.
type currentGcsEnv struct {
	Key, Secret string
}

func getGcsEnv() *currentGcsEnv {
	// Grab any existing Fastly GCS keys and preserve, in the off chance
	// they're actually set in the environment
	return &currentGcsEnv{
		Key:    os.Getenv("FASTLY_GCS_EMAIL"),
		Secret: os.Getenv("FASTLY_GCS_SECRET_KEY"),
	}
}

func TestResourceFastlyFlattenGCS(t *testing.T) {
	secretKey, err := generateKey()
	if err != nil {
		t.Errorf("failed to generate key: %s", err)
	}

	cases := []struct {
		remote []*gofastly.GCS
		local  []map[string]any
	}{
		{
			remote: []*gofastly.GCS{
				{
					Name:             gofastly.ToPointer("GCS collector"),
					User:             gofastly.ToPointer("email@example.com"),
					Bucket:           gofastly.ToPointer("bucketname"),
					SecretKey:        gofastly.ToPointer(secretKey),
					Format:           gofastly.ToPointer(LoggingGCSDefaultFormat),
					FormatVersion:    gofastly.ToPointer(2),
					Period:           gofastly.ToPointer(3600),
					GzipLevel:        gofastly.ToPointer(0),
					MessageType:      gofastly.ToPointer("classic"),
					CompressionCodec: gofastly.ToPointer("zstd"),
					AccountName:      gofastly.ToPointer("service-account"),
					ProjectID:        gofastly.ToPointer("project-id"),
					ProcessingRegion: gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":              "GCS collector",
					"user":              "email@example.com",
					"bucket_name":       "bucketname",
					"secret_key":        secretKey,
					"message_type":      "classic",
					"format":            LoggingGCSDefaultFormat,
					"format_version":    2,
					"period":            3600,
					"gzip_level":        0,
					"compression_codec": "zstd",
					"account_name":      "service-account",
					"project_id":        "project-id",
					"processing_region": "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenGCS(c.remote, nil)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}
