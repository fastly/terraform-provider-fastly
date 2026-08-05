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

func TestAccFastlyServiceLoggingNewRelicOTLP_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "authentication.token", "test-insert-key"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "region", "US"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_newrelicotlp.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_newrelicotlp.test", "id"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingNewRelicOTLP_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "region", "US"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "processing_region", "none"),
				),
			},
			{
				Config: ConfigLoggingNewRelicOTLPUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "authentication.token", "updated-insert-key"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "region", "EU"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "url", "https://otlp.eu01.nr-data.net"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "format_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingNewRelicOTLP_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPForImport(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "name", loggerName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_newrelicotlp.test"]
						if !ok {
							return fmt.Errorf("newrelic otlp resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_logging_newrelicotlp.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, loggerName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingNewRelicOTLP_clearToDefaults sets the optional
// attributes, then removes them, and verifies each reverts to its schema
// default (or, for placement, to unset — it has no default) rather than
// leaving a perpetual diff.
func TestAccFastlyServiceLoggingNewRelicOTLP_clearToDefaults(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "region", "EU"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "url", "https://otlp.eu01.nr-data.net"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "processing_region", "eu"),
				),
			},
			{
				Config: ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "region", "US"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "url", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "format_version", "2"),
					// placement is left unconfigured here, which is distinct from
					// explicitly set to "none" — see
					// TestAccFastlyServiceLoggingNewRelicOTLP_placementUnsetVsNone.
					resource.TestCheckNoResourceAttr("fastly_service_logging_newrelicotlp.test", "placement"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingNewRelicOTLP_placementUnsetVsNone verifies that
// leaving placement unconfigured and explicitly setting it to "none" are
// distinct, round-trippable states — not just "on create" but across updates
// in both directions — rather than being collapsed together, since the API
// treats an unset placement (auto-place in vcl_log/vcl_deliver) differently
// from an explicit "none" (suppress the log statement entirely).
func TestAccFastlyServiceLoggingNewRelicOTLP_placementUnsetVsNone(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				// Start unset.
				Config: ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_newrelicotlp.test", "placement"),
				),
			},
			{
				// Update to explicit "none".
				Config: ConfigLoggingNewRelicOTLPUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "placement", "none"),
				),
			},
			{
				// Update back to unset.
				Config: ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_newrelicotlp.test", "placement"),
				),
			},
			{
				// The API's null response must leave no residual diff against the
				// same, still-unset config.
				Config:   ConfigLoggingNewRelicOTLPBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingNewRelicOTLP_versionUpdateInPlace verifies that
// bumping the explicit resource's version argument is an in-place update
// against the new version rather than a destroy-and-recreate. The explicit
// clone workflow copies the endpoint into the new version, so version is
// intentionally not replacement-forcing (unlike service_id and name).
func TestAccFastlyServiceLoggingNewRelicOTLP_versionUpdateInPlace(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	var serviceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPAtVersion(serviceName, domainName, loggerName, 1),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "version", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_newrelicotlp.test"]
						if !ok {
							return fmt.Errorf("newrelic otlp resource not found")
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
				Config: ConfigLoggingNewRelicOTLPAtVersion(serviceName, domainName, loggerName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_service_logging_newrelicotlp.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "version", "2"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_newrelicotlp.test"]
						if !ok {
							return fmt.Errorf("newrelic otlp resource not found")
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
						if _, err := client.GetNewRelicOTLP(context.Background(), &fastly.GetNewRelicOTLPInput{
							ServiceID:      serviceID,
							ServiceVersion: 2,
							Name:           loggerName,
						}); err != nil {
							return fmt.Errorf("error fetching New Relic OTLP logging endpoint at version 2: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingNewRelicOTLP_computeRejectsVCLOnlyFields verifies
// that fastly_service_logging_newrelicotlp rejects format (a VCL-only
// attribute) when attached to a Compute service. The standalone resource's
// schema is shared across both service types, so this is enforced by
// ValidateNoVCLOnlyAttributesForCompute at apply time rather than by the schema
// itself.
func TestAccFastlyServiceLoggingNewRelicOTLP_computeRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigLoggingNewRelicOTLPComputeFormat(serviceName, loggerName),
				ExpectError: regexp.MustCompile("VCL-only attributes not supported on Compute services"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingNewRelicOTLP_formatDefault catches upstream changes
// to the format Fastly assigns when none is sent, which would leave
// constants.LoggingNewRelicOTLPDefaultFormat stale. Compute is used because it's
// the only path that omits format from the request - on VCL the schema default is
// always sent, so the API just echoes our own constant back.
func TestAccFastlyServiceLoggingNewRelicOTLP_formatDefault(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingNewRelicOTLPFormatDefault("fastly_service_compute.test", loggerName, 1),
				),
			},
		},
	})
}

// CheckLoggingNewRelicOTLPFormatDefault fails if the format Fastly reports for a
// logging endpoint differs from constants.LoggingNewRelicOTLPDefaultFormat. Reads
// the API directly, since ResetVCLOnlyToDefaults writes the constant into state
// without consulting the response. Only meaningful on an endpoint created without
// a format in the request.
func CheckLoggingNewRelicOTLPFormatDefault(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetNewRelicOTLP(context.Background(), &fastly.GetNewRelicOTLPInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching New Relic OTLP logging endpoint from Fastly: %w", err)
		}
		if logger == nil {
			return fmt.Errorf("New Relic OTLP logging endpoint %s not found in Fastly", loggerName)
		}

		if logger.Format == nil {
			return fmt.Errorf("Fastly returned a null format for New Relic OTLP logging endpoint %s, expected its default format", loggerName)
		}

		if got := *logger.Format; got != constants.LoggingNewRelicOTLPDefaultFormat {
			return fmt.Errorf(
				"constants.LoggingNewRelicOTLPDefaultFormat no longer matches the format Fastly assigns by default\ngot from API: %q\nconstant:     %q",
				got, constants.LoggingNewRelicOTLPDefaultFormat,
			)
		}

		return nil
	}
}

// TestAccFastlyServiceLoggingNewRelicOTLP_computeConsistentAfterApply covers the
// whole plan -> API response -> flatten -> state path on a Compute service, which
// the unit tests cannot reach. The VCL-only attributes are never sent for
// Compute, but their schema defaults still land in the plan, so the API's own
// values (a different default format, and placement forced to "none" on wasm)
// used to be read back into state and fail Terraform's post-apply consistency
// check with "Provider produced inconsistent result after apply". The trailing
// PlanOnly step then proves the same values survive a refresh with no residual
// diff.
func TestAccFastlyServiceLoggingNewRelicOTLP_computeConsistentAfterApply(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingNewRelicOTLPCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingNewRelicOTLPExistsInFastly("fastly_service_compute.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "name", loggerName),
					// The VCL-only attributes must hold their schema defaults, not
					// whatever the API returned for the wasm service.
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "format", constants.LoggingNewRelicOTLPDefaultFormat),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "response_condition", ""),
					resource.TestCheckNoResourceAttr("fastly_service_logging_newrelicotlp.test", "placement"),
				),
			},
			{
				Config:   ConfigLoggingNewRelicOTLPCompute(serviceName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// CheckLoggingNewRelicOTLPExistsInFastly verifies a New Relic OTLP logging
// endpoint exists in the Fastly API.
func CheckLoggingNewRelicOTLPExistsInFastly(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetNewRelicOTLP(context.Background(), &fastly.GetNewRelicOTLPInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching New Relic OTLP logging endpoint from Fastly: %w", err)
		}

		if logger == nil {
			return fmt.Errorf("New Relic OTLP logging endpoint %s not found in Fastly", loggerName)
		}

		return nil
	}
}
