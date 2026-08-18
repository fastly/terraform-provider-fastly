package acceptancetests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/terraform-provider-fastly/internal/constants"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccFastlyServiceLoggingGCS_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "bucket_name", "fastly-test-bucket"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", "test-gcs@fastly-test-project.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "authentication.secret_key"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "format"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "id"),
				),
			},
			{
				// The default format is a Computed default sent verbatim to the API,
				// so it must round-trip byte-for-byte and leave no residual diff.
				Config:   ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_authEnvDefaults verifies that account_name,
// email, and secret_key still pick up FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME /
// FASTLY_GCS_EMAIL / FASTLY_GCS_SECRET_KEY when the entire authentication
// object is omitted from config. This exercises the parent-level
// authenticationEnvDefault{} (schema.go): it confirms the default populates
// the omitted object from the environment rather than the object being left
// null.
//
// Not run in parallel: t.Setenv panics if the test also calls t.Parallel, and
// this test needs the env vars set for its own duration only.
func TestAccFastlyServiceLoggingGCS_authEnvDefaults(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	t.Setenv("FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME", "test-service-account")
	t.Setenv("FASTLY_GCS_ACCOUNT_NAME", "")
	t.Setenv("FASTLY_GCS_EMAIL", "")
	t.Setenv("FASTLY_GCS_SECRET_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSNoAuth(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.account_name", "test-service-account"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.secret_key", ""),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_authEnvDefaultsOwnAccountNameVar verifies
// that account_name still picks up GCS's own FASTLY_GCS_ACCOUNT_NAME
// environment variable (used by the live SDKv2 provider for this exact
// resource, and not deprecated for it — see accountNameEnvValue in
// schema.go) when FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME is unset and the entire
// authentication object is omitted from config.
//
// Not run in parallel: t.Setenv panics if the test also calls t.Parallel, and
// this test needs the env vars set for its own duration only.
func TestAccFastlyServiceLoggingGCS_authEnvDefaultsOwnAccountNameVar(t *testing.T) {
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	t.Setenv("FASTLY_GOOGLE_SERVICE_ACCOUNT_NAME", "")
	t.Setenv("FASTLY_GCS_ACCOUNT_NAME", "test-gcs-service-account")
	t.Setenv("FASTLY_GCS_EMAIL", "")
	t.Setenv("FASTLY_GCS_SECRET_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSNoAuth(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.account_name", "test-gcs-service-account"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.secret_key", ""),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingGCS_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "bucket_name", "fastly-test-bucket"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "processing_region", "none"),
				),
			},
			{
				Config: ConfigLoggingGCSUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "bucket_name", "fastly-test-bucket-updated"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", "updated-gcs@fastly-test-project.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "authentication.secret_key"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "message_type", "loggly"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "placement", "none"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_accountNameToEmailSecretKey verifies that
// switching an existing endpoint's authentication from account_name to
// email/secret_key actually clears account_name on the API side. The API
// rejects an explicit empty account_name on update, so BuildUpdateInput omits
// it — which by itself would leave the old account_name in place and diverge
// from the plan. UpdateOrRecreate handles this by deleting and recreating the
// endpoint instead.
func TestAccFastlyServiceLoggingGCS_accountNameToEmailSecretKey(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSAccountName(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.account_name", "test-service-account"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.secret_key", ""),
				),
			},
			{
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.account_name", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "authentication.email", "test-gcs@fastly-test-project.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_gcs.test", "authentication.secret_key"),
				),
			},
			{
				// The recreated endpoint's own state must leave no residual diff
				// against the same config on a subsequent refresh.
				Config:   ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceLoggingGCS_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSForImport(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "name", loggerName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_gcs.test"]
						if !ok {
							return fmt.Errorf("gcs resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_logging_gcs.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, loggerName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_clearToDefaults sets the optional attributes,
// then removes them, and verifies each reverts to its schema default (or, for
// placement, to unset — it has no default) rather than leaving a perpetual
// diff.
func TestAccFastlyServiceLoggingGCS_clearToDefaults(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "path", "/logs/"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "placement", "none"),
				),
			},
			{
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "path", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "message_type", "classic"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "response_condition", ""),
					// placement is left unconfigured here, which is distinct from
					// explicitly set to "none" — see
					// TestAccFastlyServiceLoggingGCS_placementUnsetVsNone.
					resource.TestCheckNoResourceAttr("fastly_service_logging_gcs.test", "placement"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_placementUnsetVsNone verifies that leaving
// placement unconfigured and explicitly setting it to "none" are distinct,
// round-trippable states — not just on create but across updates in both
// directions — rather than being collapsed together, since the API treats an
// unset placement (auto-place in vcl_log/vcl_deliver) differently from an
// explicit "none" (suppress the log statement entirely).
func TestAccFastlyServiceLoggingGCS_placementUnsetVsNone(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				// Start unset.
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_gcs.test", "placement"),
				),
			},
			{
				// Update to explicit "none".
				Config: ConfigLoggingGCSUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "placement", "none"),
				),
			},
			{
				// Update back to unset.
				Config: ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_gcs.test", "placement"),
				),
			},
			{
				// The API's null response must leave no residual diff against the
				// same, still-unset config.
				Config:   ConfigLoggingGCSBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_versionUpdateInPlace verifies that bumping
// the explicit resource's version argument is an in-place update against the
// new version rather than a destroy-and-recreate. The explicit clone
// workflow copies the endpoint into the new version, so version is
// intentionally not replacement-forcing (unlike service_id and name).
func TestAccFastlyServiceLoggingGCS_versionUpdateInPlace(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	var serviceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSAtVersion(serviceName, domainName, loggerName, 1),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "version", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_gcs.test"]
						if !ok {
							return fmt.Errorf("gcs resource not found")
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
				Config: ConfigLoggingGCSAtVersion(serviceName, domainName, loggerName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_service_logging_gcs.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "version", "2"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_gcs.test"]
						if !ok {
							return fmt.Errorf("gcs resource not found")
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
						if _, err := client.GetGCS(context.Background(), &fastly.GetGCSInput{
							ServiceID:      serviceID,
							ServiceVersion: 2,
							Name:           loggerName,
						}); err != nil {
							return fmt.Errorf("error fetching GCS logging endpoint at version 2: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_computeRejectsVCLOnlyFields verifies that
// fastly_service_logging_gcs rejects format (a VCL-only attribute) when
// attached to a Compute service. The standalone resource's schema is shared
// across both service types, so this is enforced by
// ValidateNoVCLOnlyAttributesForCompute at apply time rather than by the
// schema itself.
func TestAccFastlyServiceLoggingGCS_computeRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigLoggingGCSComputeFormat(serviceName, loggerName),
				ExpectError: regexp.MustCompile("VCL-only attributes not supported on Compute services"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_formatDefault catches upstream changes to
// the format Fastly assigns when none is sent, which would leave
// constants.LoggingGCSDefaultFormat stale. Compute is used because it's the
// only path that omits format from the request - on VCL the schema default is
// always sent, so the API just echoes our own constant back.
func TestAccFastlyServiceLoggingGCS_formatDefault(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingGCSFormatDefault("fastly_service_compute.test", loggerName, 1),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingGCS_computeConsistentAfterApply covers the whole
// plan -> API response -> flatten -> state path on a Compute service, which
// the unit tests cannot reach. The VCL-only attributes are never sent for
// Compute, but their schema defaults still land in the plan, so the API's own
// values (a different default format, and placement forced to "none" on
// wasm) used to be read back into state and fail Terraform's post-apply
// consistency check with "Provider produced inconsistent result after apply".
// The trailing PlanOnly step then proves the same values survive a refresh
// with no residual diff.
func TestAccFastlyServiceLoggingGCS_computeConsistentAfterApply(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("gcs-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingGCSCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingGCSExistsInFastly("fastly_service_compute.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "name", loggerName),
					// The VCL-only attributes must hold their schema defaults, not
					// whatever the API returned for the wasm service.
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format", constants.LoggingGCSDefaultFormat),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_gcs.test", "response_condition", ""),
					resource.TestCheckNoResourceAttr("fastly_service_logging_gcs.test", "placement"),
				),
			},
			{
				Config:   ConfigLoggingGCSCompute(serviceName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// CheckLoggingGCSFormatDefault fails if the format Fastly reports for a
// logging endpoint differs from constants.LoggingGCSDefaultFormat. Reads the
// API directly, since FlattenToComputeNestedModel writes the constant into
// state without consulting the response. Only meaningful on an endpoint
// created without a format in the request.
func CheckLoggingGCSFormatDefault(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetGCS(context.Background(), &fastly.GetGCSInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching GCS logging endpoint from Fastly: %w", err)
		}
		if logger == nil {
			return fmt.Errorf("GCS logging endpoint %s not found in Fastly", loggerName)
		}

		if logger.Format == nil {
			return fmt.Errorf("Fastly returned a null format for GCS logging endpoint %s, expected its default format", loggerName)
		}

		if got := *logger.Format; got != constants.LoggingGCSDefaultFormat {
			return fmt.Errorf(
				"constants.LoggingGCSDefaultFormat no longer matches the format Fastly assigns by default\ngot from API: %q\nconstant:     %q",
				got, constants.LoggingGCSDefaultFormat,
			)
		}

		return nil
	}
}

// CheckLoggingGCSExistsInFastly verifies a GCS logging endpoint exists in the
// Fastly API.
func CheckLoggingGCSExistsInFastly(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetGCS(context.Background(), &fastly.GetGCSInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching GCS logging endpoint from Fastly: %w", err)
		}

		if logger == nil {
			return fmt.Errorf("GCS logging endpoint %s not found in Fastly", loggerName)
		}

		return nil
	}
}
