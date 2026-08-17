package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCDNAuto_withHealthCheck(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	healthCheckName := fmt.Sprintf("hc_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithHealthCheck(serviceName, domainName, healthCheckName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.name", healthCheckName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.host", "example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.path", "/healthz"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.check_interval", "5000"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.expected_response", "200"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.http_version", "1.1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.initial", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.method", "HEAD"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.threshold", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.timeout", "5000"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.window", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withMultipleHealthChecks(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	healthCheckName1 := fmt.Sprintf("hc_1_%s", acctest.RandString(10))
	healthCheckName2 := fmt.Sprintf("hc_2_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleHealthChecks(serviceName, domainName, healthCheckName1, healthCheckName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.name", healthCheckName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.name", healthCheckName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.host", "other.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.path", "/status"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.check_interval", "10000"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.expected_response", "204"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.headers.#", "1"),
					resource.TestCheckTypeSetElemAttr("fastly_service_cdn_auto.test", "healthcheck.1.headers.*", "X-Api-Key: abc123"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.http_version", "1.0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.initial", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.method", "GET"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.threshold", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.timeout", "3000"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.1.window", "10"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withHealthCheckUpdate(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	healthCheckName := fmt.Sprintf("hc_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithHealthCheck(serviceName, domainName, healthCheckName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.host", "example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.threshold", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithHealthCheckUpdated(serviceName, domainName, healthCheckName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.name", healthCheckName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.host", "updated.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.path", "/status"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.check_interval", "10000"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.threshold", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.headers.#", "1"),
					resource.TestCheckTypeSetElemAttr("fastly_service_cdn_auto.test", "healthcheck.0.headers.*", "X-Api-Key: abc123"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_withHealthCheckAndBackend confirms that a backend referencing a
// health check by name applies cleanly, verifying health checks are reconciled before backend
// within the same service version (see servicecdnauto's Create/Update, which reconcile
// healthcheck immediately after condition and before backend for this reason).
func TestAccFastlyServiceCDNAuto_withHealthCheckAndBackend(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	healthCheckName := fmt.Sprintf("hc_%s", acctest.RandString(10))
	backendName := fmt.Sprintf("backend_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithHealthCheckAndBackend(serviceName, domainName, healthCheckName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "healthcheck.0.name", healthCheckName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.healthcheck", healthCheckName),
				),
			},
		},
	})
}

func TestAccFastlyServiceComputeAuto_withHealthCheck(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	healthCheckName := fmt.Sprintf("hc_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "healthcheck.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigComputeAutoWithHealthCheck(serviceName, domainName, healthCheckName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "healthcheck.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "healthcheck.0.name", healthCheckName),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "healthcheck.0.host", "example.com"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "healthcheck.0.path", "/healthz"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}
