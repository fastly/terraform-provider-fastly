package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFastlyServiceCDNAuto_withLoggingBlobStorage exercises Blob Storage logging as a
// nested block inside fastly_service_cdn_auto: adding the endpoint clones and activates a
// new version, and the reconciled state reflects the created endpoint.
func TestAccFastlyServiceCDNAuto_withLoggingBlobStorage(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingBlobStorage(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.container", containerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withLoggingBlobStorageUpdate changes optional attributes on a
// nested Blob Storage logging endpoint, exercising the reconcile update path (in-place
// update, not delete+recreate) inside a newly cloned and activated version.
func TestAccFastlyServiceCDNAuto_withLoggingBlobStorageUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithLoggingBlobStorageAll(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.container", containerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.format_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.file_max_bytes", "1048576"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingBlobStorageUpdated(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.path", "/updated-logs/"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.period", "1800"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.gzip_level", "9"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.file_max_bytes", "2097152"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withLoggingBlobStorageGzipCodec verifies that setting
// compression_codec = "gzip" on a nested endpoint (for which the API auto-populates
// gzip_level) does not produce a perpetual diff. gzip_level is left unset, so the auto
// read-back (MatchOrder/preserveGzipSentinelList) must keep it at the -1 sentinel. The
// trailing PlanOnly step fails on any residual diff.
func TestAccFastlyServiceCDNAuto_withLoggingBlobStorageGzipCodec(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithLoggingBlobStorageGzipCodec(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.compression_codec", "gzip"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.gzip_level", "-1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithLoggingBlobStorageGzipCodec(serviceName, domainName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withMultipleLoggingBlobStorage verifies that multiple nested
// Blob Storage logging endpoints reconcile correctly and preserve configured order across
// reads.
func TestAccFastlyServiceCDNAuto_withMultipleLoggingBlobStorage(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName1 := fmt.Sprintf("blobstorage-logger-1-%s", acctest.RandString(10))
	loggerName2 := fmt.Sprintf("blobstorage-logger-2-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleLoggingBlobStorage(serviceName, domainName, loggerName1, loggerName2, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName1, 1),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName2, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.name", loggerName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.1.name", loggerName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithMultipleLoggingBlobStorage(serviceName, domainName, loggerName1, loggerName2, containerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withBackendAndLoggingBlobStorage verifies Blob Storage
// logging reconciles alongside another nested block type in the same auto service.
func TestAccFastlyServiceCDNAuto_withBackendAndLoggingBlobStorage(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithBackendAndLoggingBlobStorage(serviceName, domainName, backendName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_cdn_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_blobstorage.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_withLoggingBlobStorage exercises Blob Storage logging as
// a nested block inside fastly_service_compute_auto, covering the reconcile path for the
// Compute family.
func TestAccFastlyServiceComputeAuto_withLoggingBlobStorage(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithLoggingBlobStorage(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_compute_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_blobstorage.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_blobstorage.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_blobstorage.0.container", containerName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_loggingBlobStorageRejectsVCLOnlyFields verifies that
// format (and, by extension, format_version/placement/response_condition) is not a valid
// attribute on service_compute_auto's nested logging_blobstorage block. Those attributes
// only affect generated VCL, so ComputeNestedBlockSchema omits them entirely — Terraform
// should reject this at plan time with its own "Unsupported argument" schema error, without
// ever reaching the Fastly API.
func TestAccFastlyServiceComputeAuto_loggingBlobStorageRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigComputeAutoWithLoggingBlobStorageFormat(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile(`Unsupported argument`),
			},
		},
	})
}
