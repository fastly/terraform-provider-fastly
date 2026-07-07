package acceptancetests

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceLoggingS3_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3Basic(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_s3.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_s3.test", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingS3_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3All(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "domain", "s3.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "format_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "acl", "private"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "redundancy", "standard"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "file_max_bytes", "1048576"),
				),
			},
			{
				Config: ConfigLoggingS3Updated(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "domain", "s3.eu-west-1.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "path", "/updated-logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "period", "1800"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "9"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "acl", "public-read"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "redundancy", "reduced_redundancy"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "file_max_bytes", "2097152"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingS3_iamRole(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3IAM(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "name", loggerName),
					// `iam_role` cannot be used with `access_key` + `secret_key`.
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "authentication.iam_role", "arn:aws:iam::123456789012:role/FastlyS3Access"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "authentication.access_key", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "authentication.secret_key", ""),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingS3_allAttr(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3All(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "bucket_name", bucketName),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "domain", "s3.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "format_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "acl", "private"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "redundancy", "standard"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "file_max_bytes", "1048576"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingS3_clearToDefaults(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3All(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "domain", "s3.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "acl", "private"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "redundancy", "standard"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "file_max_bytes", "1048576"),
				),
			},
			{
				Config: ConfigLoggingS3Defaults(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "domain", "s3.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "path", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "period", "3600"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "0"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "message_type", "blank"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S.000"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "acl", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "redundancy", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "file_max_bytes", "0"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingS3_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3ForImport(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "name", loggerName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_s3.test"]
						if !ok {
							return fmt.Errorf("logging s3 resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_logging_s3.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, loggerName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingS3_compressionCodec verifies that setting compression_codec
// without gzip_level does not result in an API error (the two fields are mutually exclusive).
func TestAccFastlyServiceLoggingS3_compressionCodec(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingS3CompressionCodec(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "compression_codec", "zstd"),
					resource.TestCheckResourceAttr("fastly_service_logging_s3.test", "gzip_level", "0"),
				),
			},
		},
	})
}

// CheckLoggingS3ExistsInFastly verifies an S3 logging endpoint exists in the Fastly API.
func CheckLoggingS3ExistsInFastly(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetS3(context.Background(), &fastly.GetS3Input{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching S3 logging endpoint from Fastly: %w", err)
		}

		if logger == nil {
			return fmt.Errorf("S3 logging endpoint %s not found in Fastly", loggerName)
		}

		return nil
	}
}
