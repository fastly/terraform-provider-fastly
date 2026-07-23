package acceptancetests

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "token", "test-insert-key"),
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
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "token", "updated-insert-key"),
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
// default rather than leaving a perpetual diff.
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
					resource.TestCheckResourceAttr("fastly_service_logging_newrelicotlp.test", "placement", "none"),
				),
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
