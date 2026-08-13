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

func TestAccFastlyServiceLoggingSplunk_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "url", "https://splunk.example.com/services/collector/event"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "authentication.token", "test-splunk-token"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "use_tls", "false"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_bytes", "0"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_entries", "0"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "version", "1"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "format"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "service_id"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "id"),
				),
			},
			{
				// The default format is a Computed default sent verbatim to the API,
				// so it must round-trip byte-for-byte and leave no residual diff.
				Config:   ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceLoggingSplunk_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "use_tls", "false"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "processing_region", "none"),
				),
			},
			{
				Config: ConfigLoggingSplunkUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "url", "https://splunk-updated.example.com/services/collector/event"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "authentication.token", "updated-splunk-token"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "tls.ca_cert"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "tls.client_cert"),
					resource.TestCheckResourceAttrSet("fastly_service_logging_splunk.test", "tls.client_key"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "tls.hostname", "splunk.example.com"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "use_tls", "true"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_bytes", "1000000"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_entries", "1000"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format", "%h %l %u %t \"%r\" %>s %b"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "placement", "none"),
				),
			},
		},
	})
}

func TestAccFastlyServiceLoggingSplunk_importBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	var serviceID string
	var versionNumber string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkForImport(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "name", loggerName),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_splunk.test"]
						if !ok {
							return fmt.Errorf("splunk resource not found")
						}
						serviceID = rs.Primary.Attributes["service_id"]
						versionNumber = rs.Primary.Attributes["version"]
						return nil
					},
				),
			},
			{
				ResourceName: "fastly_service_logging_splunk.test",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return fmt.Sprintf("%s/%s/%s", serviceID, versionNumber, loggerName), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_clearToDefaults sets the optional
// attributes, then removes them, and verifies each reverts to its schema default
// (or, for placement, to unset — it has no default) rather than leaving a
// perpetual diff.
func TestAccFastlyServiceLoggingSplunk_clearToDefaults(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "use_tls", "true"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "placement", "none"),
				),
			},
			{
				Config: ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "use_tls", "false"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "processing_region", "none"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_bytes", "0"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "request_max_entries", "0"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "tls.ca_cert", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "tls.client_key", ""),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "response_condition", ""),
					// placement is left unconfigured here, which is distinct from
					// explicitly set to "none" — see
					// TestAccFastlyServiceLoggingSplunk_placementUnsetVsNone.
					resource.TestCheckNoResourceAttr("fastly_service_logging_splunk.test", "placement"),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_placementUnsetVsNone verifies that leaving
// placement unconfigured and explicitly setting it to "none" are distinct,
// round-trippable states — not just "on create" but across updates in both
// directions — rather than being collapsed together, since the API treats an
// unset placement (auto-place in vcl_log/vcl_deliver) differently from an
// explicit "none" (suppress the log statement entirely).
func TestAccFastlyServiceLoggingSplunk_placementUnsetVsNone(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				// Start unset.
				Config: ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_splunk.test", "placement"),
				),
			},
			{
				// Update to explicit "none".
				Config: ConfigLoggingSplunkUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "placement", "none"),
				),
			},
			{
				// Update back to unset.
				Config: ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckNoResourceAttr("fastly_service_logging_splunk.test", "placement"),
				),
			},
			{
				// The API's null response must leave no residual diff against the
				// same, still-unset config.
				Config:   ConfigLoggingSplunkBasic(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_versionUpdateInPlace verifies that bumping
// the explicit resource's version argument is an in-place update against the new
// version rather than a destroy-and-recreate. The explicit clone workflow copies
// the endpoint into the new version, so version is intentionally not
// replacement-forcing (unlike service_id and name).
func TestAccFastlyServiceLoggingSplunk_versionUpdateInPlace(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	var serviceID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkAtVersion(serviceName, domainName, loggerName, 1),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "version", "1"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_splunk.test"]
						if !ok {
							return fmt.Errorf("splunk resource not found")
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
				Config: ConfigLoggingSplunkAtVersion(serviceName, domainName, loggerName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_service_logging_splunk.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn.test"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "version", "2"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["fastly_service_logging_splunk.test"]
						if !ok {
							return fmt.Errorf("splunk resource not found")
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
						if _, err := client.GetSplunk(context.Background(), &fastly.GetSplunkInput{
							ServiceID:      serviceID,
							ServiceVersion: 2,
							Name:           loggerName,
						}); err != nil {
							return fmt.Errorf("error fetching Splunk logging endpoint at version 2: %w", err)
						}

						return nil
					},
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_computeRejectsVCLOnlyFields verifies that
// fastly_service_logging_splunk rejects format (a VCL-only attribute) when
// attached to a Compute service. The standalone resource's schema is shared
// across both service types, so this is enforced by
// ValidateNoVCLOnlyAttributesForCompute at apply time rather than by the schema
// itself.
func TestAccFastlyServiceLoggingSplunk_computeRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigLoggingSplunkComputeFormat(serviceName, loggerName),
				ExpectError: regexp.MustCompile("VCL-only attributes not supported on Compute services"),
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_formatDefault catches upstream changes to
// the format Fastly assigns when none is sent, which would leave
// constants.LoggingSplunkDefaultFormat stale. Compute is used because it's the
// only path that omits format from the request - on VCL the schema default is
// always sent, so the API just echoes our own constant back.
func TestAccFastlyServiceLoggingSplunk_formatDefault(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingSplunkFormatDefault("fastly_service_compute.test", loggerName, 1),
				),
			},
		},
	})
}

// TestAccFastlyServiceLoggingSplunk_computeConsistentAfterApply covers the whole
// plan -> API response -> flatten -> state path on a Compute service, which the
// unit tests cannot reach. The VCL-only attributes are never sent for Compute,
// but their schema defaults still land in the plan, so the API's own values
// (a different default format, and placement forced to "none" on wasm) used to be
// read back into state and fail Terraform's post-apply consistency check with
// "Provider produced inconsistent result after apply". The trailing PlanOnly step
// then proves the same values survive a refresh with no residual diff.
func TestAccFastlyServiceLoggingSplunk_computeConsistentAfterApply(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	loggerName := fmt.Sprintf("splunk-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute"),
		Steps: []resource.TestStep{
			{
				Config: ConfigLoggingSplunkCompute(serviceName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute.test"),
					CheckLoggingSplunkExistsInFastly("fastly_service_compute.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "name", loggerName),
					// The VCL-only attributes must hold their schema defaults, not
					// whatever the API returned for the wasm service.
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format", constants.LoggingSplunkDefaultFormat),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_logging_splunk.test", "response_condition", ""),
					resource.TestCheckNoResourceAttr("fastly_service_logging_splunk.test", "placement"),
				),
			},
			{
				Config:   ConfigLoggingSplunkCompute(serviceName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// CheckLoggingSplunkFormatDefault fails if the format Fastly reports for a
// logging endpoint differs from constants.LoggingSplunkDefaultFormat. Reads the
// API directly, since FlattenToComputeNestedModel writes the constant into state
// without consulting the response. Only meaningful on an endpoint created
// without a format in the request.
func CheckLoggingSplunkFormatDefault(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetSplunk(context.Background(), &fastly.GetSplunkInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching Splunk logging endpoint from Fastly: %w", err)
		}
		if logger == nil {
			return fmt.Errorf("Splunk logging endpoint %s not found in Fastly", loggerName)
		}

		if logger.Format == nil {
			return fmt.Errorf("Fastly returned a null format for Splunk logging endpoint %s, expected its default format", loggerName)
		}

		if got := *logger.Format; got != constants.LoggingSplunkDefaultFormat {
			return fmt.Errorf(
				"constants.LoggingSplunkDefaultFormat no longer matches the format Fastly assigns by default\ngot from API: %q\nconstant:     %q",
				got, constants.LoggingSplunkDefaultFormat,
			)
		}

		return nil
	}
}

// CheckLoggingSplunkExistsInFastly verifies a Splunk logging endpoint exists in
// the Fastly API.
func CheckLoggingSplunkExistsInFastly(serviceName, loggerName string, version int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[serviceName]
		if !ok {
			return fmt.Errorf("service not found: %s", serviceName)
		}

		client, err := NewFastlyClient()
		if err != nil {
			return fmt.Errorf("error creating Fastly client: %w", err)
		}

		logger, err := client.GetSplunk(context.Background(), &fastly.GetSplunkInput{
			ServiceID:      rs.Primary.ID,
			ServiceVersion: version,
			Name:           loggerName,
		})
		if err != nil {
			return fmt.Errorf("error fetching Splunk logging endpoint from Fastly: %w", err)
		}

		if logger == nil {
			return fmt.Errorf("Splunk logging endpoint %s not found in Fastly", loggerName)
		}

		return nil
	}
}
