package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFastlyServiceCDNAuto_withLoggingS3 exercises S3 logging as a nested block
// inside fastly_service_cdn_auto: adding the endpoint clones and activates a new
// version, and the reconciled state reflects the created endpoint.
func TestAccFastlyServiceCDNAuto_withLoggingS3(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingS3(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withLoggingS3Update changes optional attributes on a
// nested S3 logging endpoint, exercising the reconcile update path (in-place update,
// not delete+recreate) inside a newly cloned and activated version.
func TestAccFastlyServiceCDNAuto_withLoggingS3Update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithLoggingS3All(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.domain", "s3.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.format_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.acl", "private"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.redundancy", "standard"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.file_max_bytes", "1048576"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingS3Updated(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.domain", "s3.eu-west-1.amazonaws.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.path", "/updated-logs/"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.period", "1800"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.gzip_level", "9"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.acl", "public-read"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.redundancy", "reduced_redundancy"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.file_max_bytes", "2097152"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withLoggingS3GzipCodec verifies that setting
// compression_codec = "gzip" on a nested endpoint (for which the API auto-populates
// gzip_level) does not produce a perpetual diff. gzip_level is left unset, so the auto
// read-back (MatchOrder/preserveGzipSentinelList) must keep it at the -1 sentinel. The
// trailing PlanOnly step fails on any residual diff.
func TestAccFastlyServiceCDNAuto_withLoggingS3GzipCodec(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithLoggingS3GzipCodec(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.compression_codec", "gzip"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.gzip_level", "-1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithLoggingS3GzipCodec(serviceName, domainName, loggerName, bucketName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withMultipleLoggingS3 verifies that multiple nested S3
// logging endpoints reconcile correctly and preserve configured order across reads.
func TestAccFastlyServiceCDNAuto_withMultipleLoggingS3(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName1 := fmt.Sprintf("s3-logger-1-%s", acctest.RandString(10))
	loggerName2 := fmt.Sprintf("s3-logger-2-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleLoggingS3(serviceName, domainName, loggerName1, loggerName2, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName1, 1),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName2, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.name", loggerName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.1.name", loggerName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithMultipleLoggingS3(serviceName, domainName, loggerName1, loggerName2, bucketName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withBackendAndLoggingS3 verifies S3 logging reconciles
// alongside another nested block type in the same auto service.
func TestAccFastlyServiceCDNAuto_withBackendAndLoggingS3(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithBackendAndLoggingS3(serviceName, domainName, backendName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_cdn_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_s3.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_withLoggingS3 exercises S3 logging as a nested block
// inside fastly_service_compute_auto, covering the reconcile path for the Compute family.
func TestAccFastlyServiceComputeAuto_withLoggingS3(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithLoggingS3(serviceName, domainName, loggerName, bucketName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					CheckLoggingS3ExistsInFastly("fastly_service_compute_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_s3.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_s3.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_s3.0.bucket_name", bucketName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_loggingS3RejectsVCLOnlyFields verifies that
// format (and, by extension, format_version/placement/response_condition) is
// not a valid attribute on service_compute_auto's nested logging_s3 block.
// Those attributes only affect generated VCL, so ComputeNestedBlockSchema
// omits them entirely — Terraform should reject this at plan time with its
// own "Unsupported argument" schema error, without ever reaching the Fastly API.
func TestAccFastlyServiceComputeAuto_loggingS3RejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("s3-logger-%s", acctest.RandString(10))
	bucketName := fmt.Sprintf("tf-test-bucket-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigComputeAutoWithLoggingS3Format(serviceName, domainName, loggerName, bucketName),
				ExpectError: regexp.MustCompile(`Unsupported argument`),
			},
		},
	})
}
