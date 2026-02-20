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

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

func TestAccFastlyServiceLoggingBlobstorage_vcl_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	blobStorageLogOne := gofastly.BlobStorage{
		AccountName:       gofastly.ToPointer("test"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Container:         gofastly.ToPointer("fastly"),
		FileMaxBytes:      gofastly.ToPointer(1048576),
		Format:            gofastly.ToPointer(LoggingBlobStorageDefaultFormat),
		FormatVersion:     gofastly.ToPointer(1),
		GzipLevel:         gofastly.ToPointer(0), // API defaults to zero
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("test-blobstorage-1"),
		Path:              gofastly.ToPointer("/5XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("error_response_5XX"),
		SASToken:          gofastly.ToPointer("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	blobStorageLogOneUpdated := gofastly.BlobStorage{
		AccountName:       gofastly.ToPointer("test"),
		Container:         gofastly.ToPointer("fastly"),
		FileMaxBytes:      gofastly.ToPointer(1048576),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(1),
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("test-blobstorage-1"),
		Path:              gofastly.ToPointer("/5XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("error_response_5XX"),
		SASToken:          gofastly.ToPointer("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	blobStorageLogTwo := gofastly.BlobStorage{
		AccountName:       gofastly.ToPointer("test"),
		CompressionCodec:  gofastly.ToPointer("zstd"),
		Container:         gofastly.ToPointer("fastly"),
		FileMaxBytes:      gofastly.ToPointer(2097152),
		Format:            gofastly.ToPointer(LoggingFormatUpdate),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0), // API defaults to zero
		MessageType:       gofastly.ToPointer("blank"),
		Name:              gofastly.ToPointer("test-blobstorage-2"),
		Path:              gofastly.ToPointer("/2XX/"),
		Period:            gofastly.ToPointer(12),
		Placement:         gofastly.ToPointer("none"),
		PublicKey:         gofastly.ToPointer(pgpPublicKey(t)),
		ResponseCondition: gofastly.ToPointer("ok_response_2XX"),
		SASToken:          gofastly.ToPointer("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
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
				Config: testAccServiceVCLBlobStorageLoggingConfigComplete(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLogOne}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_blobstorage.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLBlobStorageLoggingConfigUpdate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLogOneUpdated, &blobStorageLogTwo}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_blobstorage.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingBlobstorage_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	blobStorageLogOne := gofastly.BlobStorage{
		AccountName:      gofastly.ToPointer("test"),
		CompressionCodec: gofastly.ToPointer("zstd"),
		Container:        gofastly.ToPointer("fastly"),
		FileMaxBytes:     gofastly.ToPointer(1048576),
		GzipLevel:        gofastly.ToPointer(0), // API defaults to zero
		MessageType:      gofastly.ToPointer("blank"),
		Name:             gofastly.ToPointer("test-blobstorage-1"),
		Path:             gofastly.ToPointer("/5XX/"),
		Period:           gofastly.ToPointer(12),
		PublicKey:        gofastly.ToPointer(pgpPublicKey(t)),
		SASToken:         gofastly.ToPointer("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"),
		TimestampFormat:  gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
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
				Config: testAccServiceVCLBlobStorageLoggingConfigCompleteCompute(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLBlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLogOne}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_blobstorage.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingBlobstorage_vcl_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	blobStorageLog := gofastly.BlobStorage{
		AccountName:       gofastly.ToPointer("test"),
		Container:         gofastly.ToPointer("fastly"),
		FileMaxBytes:      gofastly.ToPointer(0),
		Format:            gofastly.ToPointer(LoggingBlobStorageDefaultFormat),
		FormatVersion:     gofastly.ToPointer(2),
		GzipLevel:         gofastly.ToPointer(0), // API defaults to zero
		MessageType:       gofastly.ToPointer("classic"),
		Name:              gofastly.ToPointer("test-blobstorage"),
		Path:              gofastly.ToPointer(""),
		Period:            gofastly.ToPointer(3600),
		PublicKey:         gofastly.ToPointer(""),
		ResponseCondition: gofastly.ToPointer(""),
		SASToken:          gofastly.ToPointer("sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"),
		TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	// FileMaxBytes Path PublicKey ResponseCondition

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLBlobStorageLoggingConfigDefault(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLBlobStorageLoggingAttributes(&service, []*gofastly.BlobStorage{&blobStorageLog}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_blobstorage.#", "1"),
				),
			},
		},
	})
}

func TestBlobstorageloggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingBlobStorage(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_blobstorage"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["sas_token"].Sensitive {
		t.Fatalf("Expected sas_token to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	token := "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%3A00%3A00Z&sig=3ABdLOJZosCp0o491T%2BqZGKIhafF1nlM3MzESDDD3Gg%3D"
	resetEnv := setBlobStorageEnv(token, t)
	defer resetEnv()

	sasTokenResult, err := loggingResourceSchema["sas_token"].DefaultFunc()
	if err != nil {
		t.Fatalf("Unexpected err %#v when calling sas_token DefaultFunc", err)
	}
	if sasTokenResult != token {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", token, sasTokenResult)
	}

	formatSchema, ok := loggingResourceSchema["format"]
	if !ok {
		t.Fatalf("Expected format field to exist in schema")
	}
	if formatSchema.Default != LoggingBlobStorageDefaultFormat {
		t.Fatalf("Error matching format default:\nexpected: %#v\ngot: %#v", LoggingBlobStorageDefaultFormat, formatSchema.Default)
	}
}

func testAccCheckFastlyServiceVCLBlobStorageLoggingAttributes(service *gofastly.ServiceDetail, localBlobStorageList []*gofastly.BlobStorage, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		remoteBlobStorageList, err := conn.ListBlobStorages(context.TODO(), &gofastly.ListBlobStoragesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Blob Storage Logging for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(remoteBlobStorageList) != len(localBlobStorageList) {
			return fmt.Errorf("blob Storage List count mismatch, expected (%d), got (%d)", len(localBlobStorageList), len(remoteBlobStorageList))
		}

		var found int
		for _, lbs := range localBlobStorageList {
			for _, rbs := range remoteBlobStorageList {
				if gofastly.ToValue(lbs.Name) == gofastly.ToValue(rbs.Name) {
					// we don't know these things ahead of time, so populate them now
					lbs.ServiceID = service.ServiceID
					lbs.ServiceVersion = service.ActiveVersion.Number

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						lbs.FormatVersion = rbs.FormatVersion
						lbs.Format = rbs.Format
						lbs.ResponseCondition = rbs.ResponseCondition
						lbs.Placement = rbs.Placement
					}

					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					rbs.CreatedAt = nil
					rbs.UpdatedAt = nil
					if diff := cmp.Diff(lbs, rbs); diff != "" {
						return fmt.Errorf("bad match Blob Storage logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(localBlobStorageList) {
			return fmt.Errorf("error matching Blob Storage Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLBlobStorageLoggingConfigComplete(serviceName string) string {
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

  logging_blobstorage {
    name = "test-blobstorage-1"
    path = "/5XX/"
    account_name = "test"
    container = "fastly"
    sas_token = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    public_key = file("test_fixtures/fastly_test_publickey")
    format_version = 1
    message_type = "blank"
    placement = "none"
    response_condition = "error_response_5XX"
    file_max_bytes     = 1048576
    compression_codec = "zstd"
    processing_region = "us"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceVCLBlobStorageLoggingConfigCompleteCompute(serviceName string) string {
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

  logging_blobstorage {
    account_name = "test"
    compression_codec = "zstd"
    container = "fastly"
    file_max_bytes     = 1048576
    message_type = "blank"
    name = "test-blobstorage-1"
    path = "/5XX/"
    period = 12
    public_key = file("test_fixtures/fastly_test_publickey")
    sas_token = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceVCLBlobStorageLoggingConfigUpdate(serviceName string) string {
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

  logging_blobstorage {
    name = "test-blobstorage-1"
    path = "/5XX/"
    account_name = "test"
    container = "fastly"
    sas_token = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = %q
    gzip_level = 1
    format_version = 2
    message_type = "blank"
    placement = "none"
    response_condition = "error_response_5XX"
    file_max_bytes     = 1048576
  }

  logging_blobstorage {
    name = "test-blobstorage-2"
    path = "/2XX/"
    account_name = "test"
    container = "fastly"
    sas_token = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
    period = 12
    timestamp_format = "%%Y-%%m-%%dT%%H:%%M:%%S.000"
    public_key = file("test_fixtures/fastly_test_publickey")
    format = %q
    format_version = 2
    message_type = "blank"
    placement = "none"
    response_condition = "ok_response_2XX"
    file_max_bytes     = 2097152
    compression_codec  = "zstd"
  }

  force_destroy = true
}`, serviceName, domainName, format, format)
}

func testAccServiceVCLBlobStorageLoggingConfigDefault(serviceName string) string {
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

  logging_blobstorage {
    name = "test-blobstorage"
    account_name = "test"
    container = "fastly"
    sas_token = "sv=2018-04-05&ss=b&srt=sco&sp=rw&se=2050-07-21T18%%3A00%%3A00Z&sig=3ABdLOJZosCp0o491T%%2BqZGKIhafF1nlM3MzESDDD3Gg%%3D"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func setBlobStorageEnv(sas string, t *testing.T) func() {
	e := getBlobStorageEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", sas); err != nil {
		t.Fatalf("Error setting env var FASTLY_AZURE_SHARED_ACCESS_SIGNATURE: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", e.SASToken); err != nil {
			t.Fatalf("Error resetting env var FASTLY_AZURE_SHARED_ACCESS_SIGNATURE: %s", err)
		}
	}
}

// struct to preserve the current environment.
type currentBlobStorageEnv struct {
	SASToken string
}

func getBlobStorageEnv() *currentBlobStorageEnv {
	// Grab the existing Fastly Azure SAS token and preserve, in the off chance
	// they're actually set in the environment
	return &currentBlobStorageEnv{
		SASToken: os.Getenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE"),
	}
}

func TestResourceFastlyFlattenBlobStorage(t *testing.T) {
	cases := []struct {
		remote []*gofastly.BlobStorage
		local  []map[string]any
	}{
		{
			remote: []*gofastly.BlobStorage{
				{
					Name:              gofastly.ToPointer("test-blobstorage"),
					Path:              gofastly.ToPointer("/logs/"),
					AccountName:       gofastly.ToPointer("test"),
					Container:         gofastly.ToPointer("fastly"),
					SASToken:          gofastly.ToPointer("test-sas-token"),
					Period:            gofastly.ToPointer(12),
					TimestampFormat:   gofastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
					GzipLevel:         gofastly.ToPointer(0),
					PublicKey:         gofastly.ToPointer("test-public-key"),
					Format:            gofastly.ToPointer(LoggingBlobStorageDefaultFormat),
					FormatVersion:     gofastly.ToPointer(2),
					MessageType:       gofastly.ToPointer("classic"),
					Placement:         gofastly.ToPointer("none"),
					ResponseCondition: gofastly.ToPointer("error_response"),
					FileMaxBytes:      gofastly.ToPointer(1048576),
					CompressionCodec:  gofastly.ToPointer("zstd"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"name":               "test-blobstorage",
					"path":               "/logs/",
					"account_name":       "test",
					"container":          "fastly",
					"sas_token":          "test-sas-token",
					"period":             12,
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
					"public_key":         "test-public-key",
					"format":             LoggingBlobStorageDefaultFormat,
					"format_version":     2,
					"gzip_level":         0,
					"message_type":       "classic",
					"placement":          "none",
					"response_condition": "error_response",
					"file_max_bytes":     1048576,
					"compression_codec":  "zstd",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBlobStorages(c.remote, nil)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
