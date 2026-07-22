package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFastlyServiceCDNAuto_withLoggingNewRelicOTLP exercises New Relic OTLP
// logging as a nested block inside fastly_service_cdn_auto: adding the endpoint
// clones and activates a new version, and the reconciled state reflects the
// created endpoint.
func TestAccFastlyServiceCDNAuto_withLoggingNewRelicOTLP(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingNewRelicOTLP(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingNewRelicOTLPExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.token", "test-insert-key"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withLoggingNewRelicOTLPUpdate changes optional
// attributes on a nested New Relic OTLP logging endpoint, exercising the
// reconcile update path (in-place update, not delete+recreate) inside a newly
// cloned and activated version.
func TestAccFastlyServiceCDNAuto_withLoggingNewRelicOTLPUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithLoggingNewRelicOTLP(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingNewRelicOTLPExistsInFastly("fastly_service_cdn_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.region", "US"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithLoggingNewRelicOTLPUpdated(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					CheckLoggingNewRelicOTLPExistsInFastly("fastly_service_cdn_auto.test", loggerName, 2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.token", "updated-insert-key"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.region", "EU"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.processing_region", "eu"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "logging_newrelicotlp.0.format_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				Config:   ConfigCDNAutoWithLoggingNewRelicOTLPUpdated(serviceName, domainName, loggerName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_withLoggingNewRelicOTLP exercises New Relic
// OTLP logging as a nested block inside fastly_service_compute_auto, covering
// the reconcile path for the Compute family.
func TestAccFastlyServiceComputeAuto_withLoggingNewRelicOTLP(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoWithLoggingNewRelicOTLP(serviceName, domainName, loggerName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					CheckLoggingNewRelicOTLPExistsInFastly("fastly_service_compute_auto.test", loggerName, 1),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_newrelicotlp.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_newrelicotlp.0.name", loggerName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "logging_newrelicotlp.0.token", "test-insert-key"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_loggingNewRelicOTLPRejectsVCLOnlyFields
// verifies that format (and, by extension,
// format_version/placement/response_condition) is not a valid attribute on
// service_compute_auto's nested logging_newrelicotlp block. Those attributes
// only affect generated VCL, so ComputeNestedBlockSchema omits them entirely —
// Terraform should reject this at plan time with its own "Unsupported argument"
// schema error, without ever reaching the Fastly API.
func TestAccFastlyServiceComputeAuto_loggingNewRelicOTLPRejectsVCLOnlyFields(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	loggerName := fmt.Sprintf("newrelic-logger-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigComputeAutoWithLoggingNewRelicOTLPFormat(serviceName, domainName, loggerName),
				ExpectError: regexp.MustCompile(`Unsupported argument`),
			},
		},
	})
}
