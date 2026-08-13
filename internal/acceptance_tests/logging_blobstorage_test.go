package acceptancetests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/fastly/terraform-provider-fastly/internal/constants"
)

func TestAccFastlyServiceLoggingBlobStorage_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageBasic(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "container", containerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_blobstorage.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_blobstorage.test", "id"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_authEnvDefaults verifies that account_name and
// sas_token still pick up FASTLY_AZURE_ACCOUNT_NAME / FASTLY_AZURE_SHARED_ACCESS_SIGNATURE
// when the entire authentication object is omitted from config. This exercises the
// parent-level authenticationEnvDefault{} (schema.go): it confirms the default populates
// the omitted object from the environment rather than the object being left null.
//
// Not run in parallel: t.Setenv panics if the test also calls t.Parallel, and this test
// needs the env vars set for its own duration only.
func TestAccFastlyServiceLoggingBlobStorage_authEnvDefaults(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	t.Setenv("FASTLY_AZURE_ACCOUNT_NAME", "envaccount")
	t.Setenv("FASTLY_AZURE_SHARED_ACCESS_SIGNATURE", "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageNoAuth(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.account_name", "envaccount"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.sas_token", "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_missingAuthentication verifies that configuring
// only one of account_name/sas_token (with no environment variable fallback for the other)
// fails at plan time via authenticationRequired, rather than producing an unclear apply-time
// error from the Fastly API.
func TestAccFastlyServiceLoggingBlobStorage_missingAuthentication(t *testing.T) {
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
				Config:      ConfigLoggingBlobStorageMissingAuth(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile("must set both `account_name` and `sas_token`"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_fileMaxBytesRange verifies that a non-zero
// file_max_bytes below the Fastly API's 1 MiB minimum fails at plan time via
// fileMaxBytesRange, rather than at apply time.
func TestAccFastlyServiceLoggingBlobStorage_fileMaxBytesRange(t *testing.T) {
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
				Config:      ConfigLoggingBlobStorageFileMaxBytesInvalid(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile("file_max_bytes must be 0"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_gzipLevelRange verifies that an explicitly
// configured gzip_level outside 0-9 fails at plan time via int64validator.Between,
// rather than at apply time.
func TestAccFastlyServiceLoggingBlobStorage_gzipLevelRange(t *testing.T) {
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
				Config:      ConfigLoggingBlobStorageGzipLevelInvalid(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile("value must be between 0 and 9"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_gzipLevelSentinelRejected verifies that
// explicitly configuring gzip_level = -1 (the internal "unset" sentinel) is
// rejected at plan time, rather than being silently accepted and reinterpreted as
// "unset" - a user should omit the attribute to get that behavior.
func TestAccFastlyServiceLoggingBlobStorage_gzipLevelSentinelRejected(t *testing.T) {
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
				Config:      ConfigLoggingBlobStorageGzipLevelSentinel(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile("value must be between 0 and 9"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_allAttr exercises every attribute of the
// standalone resource across a create and an update, so a regression in any single
// setter (e.g. sending an empty value via a Nullable helper that omits rather than
// clears a field, as opposed to new()) shows up as a wrong post-apply value here
// rather than only in an untested corner. authentication.account_name/sas_token are
// rotated between steps to exercise the credential-clearing path in
// buildCommonUpdateInput; format is deliberately omitted from the update config to
// confirm it falls back to its schema Default rather than retaining the prior value.
func TestAccFastlyServiceLoggingBlobStorage_allAttr(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageAll(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "container", containerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.account_name", "teststorageaccount"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.sas_token", "sv=2020-09-05&sr=b&sig=Z%2FRHIX5Xcg0Mq2rqI3OlWTjEg2tYkboXr1P9ZUXDtkk%3D&se=2050-09-30T02%3A23%3A26Z&sp=rw"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "file_max_bytes", "1048576"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_blobstorage.test", "public_key"),
				),
			},
			{
				Config: ConfigLoggingBlobStorageUpdated(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.account_name", "teststorageaccount2"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "authentication.sas_token", "sv=2021-08-06&sr=b&sig=A%2Fx8h5vQ3ZuTn2R9tYkX7wL0mCq1oPzB9dFsEjKa4Uc%3D&se=2051-01-01T00%3A00%3A00Z&sp=rw"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "path", "/updated-logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "period", "1800"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "9"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format", constants.LoggingBlobStorageDefaultFormat),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "file_max_bytes", "2097152"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_blobstorage.test", "public_key"),
				),
			},
			{
				// The update must leave no residual diff against the same config.
				Config:   ConfigLoggingBlobStorageUpdated(serviceName, domainName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceLoggingBlobStorage_clearToDefaults(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageAll(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "period", "7200"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "5"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S%z"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "processing_region", "us"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "file_max_bytes", "1048576"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_blobstorage.test", "public_key"),
				),
			},
			{
				Config: ConfigLoggingBlobStorageDefaults(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					// public_key must actually clear to "" on update, not be left in
					// place by a Nullable setter that omits an empty value instead of
					// sending it.
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "public_key", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "path", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "period", "3600"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "-1"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "timestamp_format", "%Y-%m-%dT%H:%M:%S.000"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "file_max_bytes", "0"),
					// placement has no default: unconfigured means unset, distinct
					// from "none" — see TestAccFastlyServiceLoggingBlobStorage_placementUnsetVsNone.
					resource.TestCheckNoResourceAttr("fastly_service_logging_blobstorage.test", "placement"),
				),
			},
			{
				// The clear must leave no residual diff against the same config.
				Config:   ConfigLoggingBlobStorageDefaults(serviceName, domainName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_placementUnsetVsNone verifies that leaving
// placement unconfigured and explicitly setting it to "none" are distinct, round-trippable
// states — not just on create but across updates in both directions — rather than being
// collapsed together, since the API treats an unset placement (auto-place in
// vcl_log/vcl_deliver) differently from an explicit "none" (suppress the log statement
// entirely).
func TestAccFastlyServiceLoggingBlobStorage_placementUnsetVsNone(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				// Start unset.
				Config: ConfigLoggingBlobStorageBasic(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_blobstorage.test", "placement"),
				),
			},
			{
				// Update to explicit "none".
				Config: ConfigLoggingBlobStoragePlacementNone(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "placement", "none"),
				),
			},
			{
				// Update back to unset.
				Config: ConfigLoggingBlobStorageBasic(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_blobstorage.test", "placement"),
				),
			},
			{
				// The API's null response must leave no residual diff against the
				// same, still-unset config.
				Config:   ConfigLoggingBlobStorageBasic(serviceName, domainName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceLoggingBlobStorage_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageForImport(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_blobstorage.test"]
						if !ok {
							return fmt.Errorf("logging blobstorage resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_logging_blobstorage.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, loggerName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_compressionCodec verifies that setting
// compression_codec without gzip_level does not result in an API error (the two
// fields are mutually exclusive). With gzip_level unset it stays at the -1
// sentinel and is never sent to the API. Also verifies compression_codec clears
// back to its "" default on update: compression_codec is Optional+Computed with a
// static "" default, and BuildUpdateInput must always send it via new() rather
// than fastly.NullString, or clearing a previously-set value never reaches the
// API (it gets omitted instead of sent as "") and the second step's check fails
// with the old value still applied.
func TestAccFastlyServiceLoggingBlobStorage_compressionCodec(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageCompressionCodec(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "compression_codec", "zstd"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "-1"),
				),
			},
			{
				Config: ConfigLoggingBlobStorageDefaults(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "compression_codec", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "-1"),
				),
			},
			{
				// The clear must leave no residual diff against the same config.
				Config:   ConfigLoggingBlobStorageDefaults(serviceName, domainName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_gzipCodec verifies that the "gzip" codec
// (for which the API auto-populates gzip_level, e.g. to 3) does not produce a
// perpetual diff. gzip_level is left unset, so it must stay at the -1 sentinel
// rather than picking up the API-managed value. The final implicit plan check
// fails on any residual diff.
func TestAccFastlyServiceLoggingBlobStorage_gzipCodec(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageGzipCodec(serviceName, domainName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "compression_codec", "gzip"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "gzip_level", "-1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_codecConflict verifies that configuring both
// compression_codec and gzip_level fails at plan time with a clear validation error,
// rather than producing an inconsistent-result error at apply time.
func TestAccFastlyServiceLoggingBlobStorage_codecConflict(t *testing.T) {
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
				Config:      ConfigLoggingBlobStorageCodecConflict(serviceName, domainName, loggerName, containerName),
				ExpectError: regexp.MustCompile("cannot be set together"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_computeRejectsVCLOnlyFields verifies that
// fastly_service_logging_blobstorage rejects format (a VCL-only attribute) when attached to
// a Compute service. The standalone resource's schema is shared across both service types,
// so this is enforced by ValidateNoVCLOnlyAttributesForCompute at apply time rather than by
// the schema itself.
func TestAccFastlyServiceLoggingBlobStorage_computeRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigLoggingBlobStorageComputeFormat(serviceName, loggerName, containerName),
				ExpectError: regexp.MustCompile("VCL-only attributes not supported on Compute services"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingBlobStorage_formatDefault catches upstream changes to the
// format Fastly assigns when none is sent, which would leave
// constants.LoggingBlobStorageDefaultFormat stale. Compute is used because it's the only
// path that omits format from the request - on VCL the schema default is always sent, so
// the API just echoes our own constant back.
func TestAccFastlyServiceLoggingBlobStorage_formatDefault(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageCompute(serviceName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingBlobStorageFormatDefault("fastly_service_compute.test", loggerName, 1),
				),
			},
		},
	})
}

// CheckLoggingBlobStorageFormatDefault fails if the format Fastly reports for a logging
// endpoint differs from constants.LoggingBlobStorageDefaultFormat. Reads the API directly,
// since ResetVCLOnlyToDefaults writes the constant into state without consulting the
// response. Only meaningful on an endpoint created without a format in the request.
func CheckLoggingBlobStorageFormatDefault(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetBlobStorage(context.Background(), &fastly.GetBlobStorageInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching Blob Storage logging endpoint from Fastly: %w", err)
		}
		if logger == nil {
			return fmt.Errorf("Blob Storage logging endpoint %s not found in Fastly", loggerName)
		}

		if logger.Format == nil {
			return fmt.Errorf("Fastly returned a null format for Blob Storage logging endpoint %s, expected its default format", loggerName)
		}

		if got := *logger.Format; got != constants.LoggingBlobStorageDefaultFormat {
			return fmt.Errorf(
				"constants.LoggingBlobStorageDefaultFormat no longer matches the format Fastly assigns by default\ngot from API: %q\nconstant:     %q",
				got, constants.LoggingBlobStorageDefaultFormat,
			)
		}

		return nil
	}
}

// TestAccFastlyServiceLoggingBlobStorage_computeConsistentAfterApply covers the whole
// plan -> API response -> flatten -> state path on a Compute service, which the unit tests
// cannot reach. The VCL-only attributes are never sent for Compute, but their schema
// defaults still land in the plan, so the API's own values (a different default format,
// and placement forced to "none" on wasm) used to be read back into state and fail
// Terraform's post-apply consistency check with "Provider produced inconsistent result
// after apply". The trailing PlanOnly step then proves the same values survive a refresh
// with no residual diff.
func TestAccFastlyServiceLoggingBlobStorage_computeConsistentAfterApply(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageCompute(serviceName, loggerName, containerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingBlobStorageExistsInFastly("fastly_service_compute.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "container", containerName),
					// The VCL-only attributes must hold their schema defaults, not
					// whatever the API returned for the wasm service.
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format", constants.LoggingBlobStorageDefaultFormat),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "response_condition", ""),
					resource.TestCheckNoResourceAttr("fastly_service_logging_blobstorage.test", "placement"),
				),
			},
			{
				Config:   ConfigLoggingBlobStorageCompute(serviceName, loggerName, containerName),
				PlanOnly: true,
			},
		},
	})
}

// CheckLoggingBlobStorageExistsInFastly verifies a Blob Storage logging endpoint exists in
// the Fastly API.
func CheckLoggingBlobStorageExistsInFastly(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetBlobStorage(context.Background(), &fastly.GetBlobStorageInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching Blob Storage logging endpoint from Fastly: %w", err)
		}

		if logger == nil {
			return fmt.Errorf("Blob Storage logging endpoint %s not found in Fastly", loggerName)
		}

		return nil
	}
}

// TestAccFastlyServiceLoggingBlobStorage_versionUpdateInPlace verifies that bumping the
// explicit resource's version argument is an in-place update against the new version
// rather than a destroy-and-recreate. The explicit clone workflow copies the endpoint into
// the new version, so version is intentionally not replacement-forcing (unlike service_id
// and name).
func TestAccFastlyServiceLoggingBlobStorage_versionUpdateInPlace(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("blobstorage-logger-%s", acctest.RandString(10))
	containerName := fmt.Sprintf("tf-test-container-%s", acctest.RandString(10))

	var serviceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingBlobStorageAtVersion(serviceName, domainName, loggerName, containerName, 1),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "version", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_blobstorage.test"]
						if !ok {
							return fmt.Errorf("logging blobstorage resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						return nil
					},
				),
			},
			{
				PreConfig: func() {
					client, err := NewFastlyClient()
					if err != nil {
						t.Fatalf("error creating Fastly client: %s", err)
					}
					if _, err := client.CloneVersion(context.Background(), &fastly.CloneVersionInput{
						ServiceID:      serviceID,
						ServiceVersion: 1,
					}); err != nil {
						t.Fatalf("error cloning version 1: %s", err)
					}
				},
				Config: ConfigLoggingBlobStorageAtVersion(serviceName, domainName, loggerName, containerName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_service_logging_blobstorage.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_blobstorage.test", "version", "2"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_blobstorage.test"]
						if !ok {
							return fmt.Errorf("logging blobstorage resource not found")
						}

						gotID := rs.Primary.Attributes["id"]
						wantID := fmt.Sprintf("%s-2-%s", serviceID, loggerName)
						if gotID != wantID {
							return fmt.Errorf("expected id %q to reflect version 2, got %q", wantID, gotID)
						}

						client, err := NewFastlyClient()
						if err != nil {
							return fmt.Errorf("error creating Fastly client: %w", err)
						}
						if _, err := client.GetBlobStorage(context.Background(), &fastly.GetBlobStorageInput{
							ServiceID:      serviceID,
							ServiceVersion: 2,
							Name:           loggerName,
						}); err != nil {
							return fmt.Errorf("error fetching Blob Storage logging endpoint at version 2: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}
