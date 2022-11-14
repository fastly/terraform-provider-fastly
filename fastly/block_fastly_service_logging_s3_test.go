package fastly

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	testAwsPrimaryAccessKey = "KEYABCDEFGHIJKLMNOPQ"
	testAwsPrimarySecretKey = "SECRET0123456789012345678901234567890123"
)

const testS3IAMRole = "arn:aws:iam::123456789012:role/S3Access"

func TestResourceFastlyFlattenS3(t *testing.T) {
	cases := []struct {
		remote []*gofastly.S3
		local  []map[string]any
	}{
		{
			remote: []*gofastly.S3{
				{
					Name:                         "s3-endpoint",
					BucketName:                   "bucket",
					Domain:                       "domain",
					AccessKey:                    testAwsPrimaryAccessKey,
					SecretKey:                    testAwsPrimarySecretKey,
					Path:                         "/",
					Period:                       3600,
					GzipLevel:                    0,
					Format:                       "%h %l %u %t %r %>s",
					FormatVersion:                2,
					ResponseCondition:            "response_condition_test",
					MessageType:                  "classic",
					TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
					Placement:                    "none",
					PublicKey:                    pgpPublicKey(t),
					Redundancy:                   "reduced_redundancy",
					ServerSideEncryptionKMSKeyID: "kmskey",
					ServerSideEncryption:         gofastly.S3ServerSideEncryptionAES,
					CompressionCodec:             "zstd",
				},
			},
			local: []map[string]any{
				{
					"name":                              "s3-endpoint",
					"bucket_name":                       "bucket",
					"domain":                            "domain",
					"s3_access_key":                     testAwsPrimaryAccessKey,
					"s3_secret_key":                     testAwsPrimarySecretKey,
					"path":                              "/",
					"period":                            3600,
					"gzip_level":                        0,
					"format":                            "%h %l %u %t %r %>s",
					"format_version":                    2,
					"response_condition":                "response_condition_test",
					"message_type":                      "classic",
					"timestamp_format":                  "%Y-%m-%dT%H:%M:%S.000",
					"placement":                         "none",
					"public_key":                        pgpPublicKey(t),
					"redundancy":                        gofastly.S3RedundancyReduced,
					"server_side_encryption":            gofastly.S3ServerSideEncryptionAES,
					"server_side_encryption_kms_key_id": "kmskey",
					"compression_codec":                 "zstd",
					"acl":                               gofastly.S3AccessControlList(""),
				},
			},
		},
		{
			remote: []*gofastly.S3{
				{
					Name:                         "s3-endpoint",
					BucketName:                   "bucket",
					Domain:                       "domain",
					IAMRole:                      testS3IAMRole,
					Path:                         "/",
					Period:                       3600,
					GzipLevel:                    5,
					Format:                       "%h %l %u %t %r %>s",
					FormatVersion:                2,
					ResponseCondition:            "response_condition_test",
					MessageType:                  "classic",
					TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
					Placement:                    "none",
					PublicKey:                    pgpPublicKey(t),
					Redundancy:                   "reduced_redundancy",
					ServerSideEncryptionKMSKeyID: "kmskey",
					ServerSideEncryption:         gofastly.S3ServerSideEncryptionAES,
					ACL:                          gofastly.S3AccessControlListPrivate,
				},
			},
			local: []map[string]any{
				{
					"name":                              "s3-endpoint",
					"bucket_name":                       "bucket",
					"domain":                            "domain",
					"s3_iam_role":                       testS3IAMRole,
					"path":                              "/",
					"period":                            3600,
					"gzip_level":                        5,
					"format":                            "%h %l %u %t %r %>s",
					"format_version":                    2,
					"response_condition":                "response_condition_test",
					"message_type":                      "classic",
					"timestamp_format":                  "%Y-%m-%dT%H:%M:%S.000",
					"placement":                         "none",
					"public_key":                        pgpPublicKey(t),
					"redundancy":                        gofastly.S3RedundancyReduced,
					"server_side_encryption":            gofastly.S3ServerSideEncryptionAES,
					"server_side_encryption_kms_key_id": "kmskey",
					"acl":                               gofastly.S3AccessControlListPrivate,
				},
			},
		},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := flattenS3s(c.remote, nil)
			if diff := cmp.Diff(out, c.local); diff != "" {
				t.Fatalf("Error matching:%s", diff)
			}
		})
	}
}

func TestAccFastlyServiceVCL_s3logging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.S3{
		ServiceVersion:    1,
		Name:              "somebucketlog",
		BucketName:        "fastlytestlogging",
		Domain:            "s3-us-west-2.amazonaws.com",
		AccessKey:         testAwsPrimaryAccessKey,
		SecretKey:         testAwsPrimarySecretKey,
		Period:            3600,
		PublicKey:         pgpPublicKey(t),
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		ResponseCondition: "response_condition_test",
		CompressionCodec:  "zstd",
	}

	log1AfterUpdate := gofastly.S3{
		ServiceVersion:    1,
		Name:              "somebucketlog",
		BucketName:        "fastlytestlogging",
		Domain:            "s3-us-west-2.amazonaws.com",
		IAMRole:           testS3IAMRole,
		Period:            3600,
		PublicKey:         pgpPublicKey(t),
		GzipLevel:         3,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "blank",
		Redundancy:        "reduced_redundancy",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		ResponseCondition: "response_condition_test",
		ACL:               gofastly.S3AccessControlListAWSExecRead,
	}

	log2 := gofastly.S3{
		ServiceVersion:   1,
		Name:             "someotherbucketlog",
		BucketName:       "fastlytestlogging2",
		Domain:           "s3-us-west-2.amazonaws.com",
		IAMRole:          testS3IAMRole,
		GzipLevel:        0,
		Period:           60,
		Format:           `%h %l %u %t "%r" %>s %b`,
		FormatVersion:    2,
		MessageType:      "classic",
		TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
		CompressionCodec: "zstd",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLS3LoggingConfig(name, domainName1, testAwsPrimaryAccessKey, testAwsPrimarySecretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLS3LoggingAttributes(&service, []*gofastly.S3{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_s3.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLS3LoggingConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLS3LoggingAttributes(&service, []*gofastly.S3{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_s3.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_s3logging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.S3{
		ServiceVersion:   1,
		Name:             "somebucketlog",
		BucketName:       "fastlytestlogging",
		Domain:           "s3-us-west-2.amazonaws.com",
		AccessKey:        testAwsPrimaryAccessKey,
		SecretKey:        testAwsPrimarySecretKey,
		Period:           3600,
		PublicKey:        pgpPublicKey(t),
		GzipLevel:        0,
		MessageType:      "classic",
		TimestampFormat:  "%Y-%m-%dT%H:%M:%S.000",
		CompressionCodec: "zstd",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLS3LoggingComputeConfig(name, domainName1, testAwsPrimaryAccessKey, testAwsPrimarySecretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLS3LoggingAttributes(&service, []*gofastly.S3{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_s3.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_s3logging_domain_default(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.S3{
		ServiceVersion:    1,
		Name:              "somebucketlog",
		BucketName:        "fastlytestlogging",
		Domain:            "s3.amazonaws.com",
		AccessKey:         testAwsPrimaryAccessKey,
		SecretKey:         testAwsPrimarySecretKey,
		Period:            3600,
		GzipLevel:         0,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		MessageType:       "classic",
		ResponseCondition: "response_condition_test",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLS3LoggingConfigDomainDefault(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLS3LoggingAttributes(&service, []*gofastly.S3{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_s3.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_s3logging_formatVersion(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.S3{
		ServiceVersion:  1,
		Name:            "somebucketlog",
		BucketName:      "fastlytestlogging",
		Domain:          "s3-us-west-2.amazonaws.com",
		AccessKey:       testAwsPrimaryAccessKey,
		SecretKey:       testAwsPrimarySecretKey,
		Period:          3600,
		GzipLevel:       0,
		Format:          "%a %l %u %t %m %U%q %H %>s %b %T",
		FormatVersion:   2,
		MessageType:     "classic",
		TimestampFormat: "%Y-%m-%dT%H:%M:%S.000",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLS3LoggingConfigFormatVersion(name, domainName1, testAwsPrimaryAccessKey, testAwsPrimarySecretKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLS3LoggingAttributes(&service, []*gofastly.S3{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_s3.#", "1"),
				),
			},
		},
	})
}

func TestS3loggingEnvDefaultFuncAttributes(t *testing.T) {
	serviceAttributes := ServiceMetadata{ServiceTypeVCL}
	v := NewServiceLoggingS3(serviceAttributes)
	r := &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
	err := v.Register(r)
	if err != nil {
		t.Fatal("Failed to register resource into schema")
	}
	loggingResource := r.Schema["logging_s3"]
	loggingResourceSchema := loggingResource.Elem.(*schema.Resource).Schema

	// Expect attributes to be sensitive
	if !loggingResourceSchema["s3_access_key"].Sensitive {
		t.Fatalf("Expected s3_access_key to be marked as a Sensitive value")
	}

	if !loggingResourceSchema["s3_secret_key"].Sensitive {
		t.Fatalf("Expected s3_secret_key to be marked as a Sensitive value")
	}

	// Actually set env var and expect it to be used to determine the values
	resetEnv := setEnv(testAwsPrimaryAccessKey, testAwsPrimarySecretKey, t)
	defer resetEnv()

	result1, err1 := loggingResourceSchema["s3_access_key"].DefaultFunc()
	if err1 != nil {
		t.Fatalf("Unexpected err %#v when calling s3_access_key DefaultFunc", err1)
	}
	if result1 != testAwsPrimaryAccessKey {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", testAwsPrimaryAccessKey, result1)
	}

	result2, err2 := loggingResourceSchema["s3_secret_key"].DefaultFunc()
	if err2 != nil {
		t.Fatalf("Unexpected err %#v when calling s3_secret_key DefaultFunc", err2)
	}
	if result2 != testAwsPrimarySecretKey {
		t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", testAwsPrimarySecretKey, result2)
	}
}

func testAccCheckFastlyServiceVCLS3LoggingAttributes(service *gofastly.ServiceDetail, s3s []*gofastly.S3, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		s3List, err := conn.ListS3s(&gofastly.ListS3sInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up S3 Logging for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(s3List) != len(s3s) {
			return fmt.Errorf("s3 List count mismatch, expected (%d), got (%d)", len(s3s), len(s3List))
		}

		var found int
		for _, s := range s3s {
			for _, ls := range s3List {
				if s.Name == ls.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					ls.CreatedAt = nil
					ls.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						ls.FormatVersion = s.FormatVersion
						ls.Format = s.Format
						ls.ResponseCondition = s.ResponseCondition
						ls.Placement = s.Placement
					}

					if diff := cmp.Diff(s, ls); diff != "" {
						return fmt.Errorf("bad match S3 logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(s3s) {
			return fmt.Errorf("error matching S3 Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLS3LoggingConfigDomainDefault(name, domain string) string {
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

  condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

  logging_s3 {
    name               = "somebucketlog"
    bucket_name        = "fastlytestlogging"
    s3_access_key      = "%s"
    s3_secret_key      = "%s"
    response_condition = "response_condition_test"
  }

  force_destroy = true
}`, name, domain, testAwsPrimaryAccessKey, testAwsPrimarySecretKey)
}

func testAccServiceVCLS3LoggingComputeConfig(name, domain, key, secret string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_s3 {
    name = "somebucketlog"
    bucket_name = "fastlytestlogging"
    domain = "s3-us-west-2.amazonaws.com"
    s3_access_key = "%s"
    s3_secret_key = "%s"
    public_key = file("test_fixtures/fastly_test_publickey")
    compression_codec = "zstd"
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }

  force_destroy = true
}`, name, domain, key, secret)
}

func testAccServiceVCLS3LoggingConfig(name, domain, key, secret string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_s3 {
    name = "somebucketlog"
    bucket_name = "fastlytestlogging"
    domain = "s3-us-west-2.amazonaws.com"
    s3_access_key = "%s"
    s3_secret_key = "%s"
    response_condition = "response_condition_test"
    public_key = file("test_fixtures/fastly_test_publickey")
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain, key, secret)
}

func testAccServiceVCLS3LoggingConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  condition {
    name = "response_condition_test"
    type = "RESPONSE"
    priority = 8
    statement = "resp.status == 418"
  }

  logging_s3 {
    name = "somebucketlog"
    bucket_name = "fastlytestlogging"
    domain = "s3-us-west-2.amazonaws.com"
    s3_iam_role = "%s"
    response_condition = "response_condition_test"
    message_type = "blank"
    public_key = file("test_fixtures/fastly_test_publickey")
    redundancy = "reduced_redundancy"
	acl = "aws-exec-read"
    gzip_level = 3
  }

  logging_s3 {
    name = "someotherbucketlog"
    bucket_name = "fastlytestlogging2"
    domain = "s3-us-west-2.amazonaws.com"
    s3_iam_role = "%s"
    period = 60
    compression_codec = "zstd"
  }

  force_destroy = true
}`, name, domain, testS3IAMRole, testS3IAMRole)
}

func testAccServiceVCLS3LoggingConfigFormatVersion(name, domain, key, secret string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name = "amazon docs"
  }

  logging_s3 {
    name = "somebucketlog"
    bucket_name = "fastlytestlogging"
    domain = "s3-us-west-2.amazonaws.com"
    s3_access_key = "%s"
    s3_secret_key = "%s"
    format = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
    format_version = 2
  }

  force_destroy = true
}`, name, domain, key, secret)
}

// struct to preserve the current environment
type currentEnv struct {
	Key, Secret string
}

func setEnv(key, secret string, t *testing.T) func() {
	e := getEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_S3_ACCESS_KEY", key); err != nil {
		t.Fatalf("Error setting env var FASTLY_S3_ACCESS_KEY: %s", err)
	}
	if err := os.Setenv("FASTLY_S3_SECRET_KEY", secret); err != nil {
		t.Fatalf("Error setting env var FASTLY_S3_SECRET_KEY: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_S3_ACCESS_KEY", e.Key); err != nil {
			t.Fatalf("Error resetting env var AWS_ACCESS_KEY_ID: %s", err)
		}
		if err := os.Setenv("FASTLY_S3_SECRET_KEY", e.Secret); err != nil {
			t.Fatalf("Error resetting env var FASTLY_S3_SECRET_KEY: %s", err)
		}
	}
}

func getEnv() *currentEnv {
	// Grab any existing Fastly AWS S3 keys and preserve, in the off chance
	// they're actually set in the environment
	return &currentEnv{
		Key:    os.Getenv("FASTLY_S3_ACCESS_KEY"),
		Secret: os.Getenv("FASTLY_S3_SECRET_KEY"),
	}
}
