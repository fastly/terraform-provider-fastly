package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFastlyServiceCDNAuto_basic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "force_destroy", "true"),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "id"),

					// Prove version 1 is bootstrapped and activated
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withBackend(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainName),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithBackend(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.address", "api.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.port", "443"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.use_ssl", "true"),
					// Adding backend should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_update(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	serviceNameUpdated := fmt.Sprintf("tf-test-updated-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	domainNameUpdated := fmt.Sprintf("%s-updated.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainName),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceNameUpdated, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainName),
					// Service name update does not create a new version (service-level attribute)
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceNameUpdated, domainNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceNameUpdated),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainNameUpdated),
					// Domain update triggers new version creation
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_multipleBackends(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "0"),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoMultipleBackends(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", "backend-primary"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.1.name", "backend-secondary"),
					// Both backend additions should land in version 2 (proves multiple nested changes in same version)
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_preservesBackendAndDomainOrder(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainBName := fmt.Sprintf("b-%s.example.com", acctest.RandString(10))
	domainAName := fmt.Sprintf("a-%s.example.com", acctest.RandString(10))
	config := ConfigCDNAutoUnsortedBackendAndDomainBlocks(serviceName, domainBName, domainAName)
	reversedConfig := ConfigCDNAutoReversedBackendAndDomainBlocks(serviceName, domainBName, domainAName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", "b"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.address", "b.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.1.name", "a"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.1.address", "a.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainBName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.1.name", domainAName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				Config: reversedConfig,
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", "a"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.address", "a.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.1.name", "b"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.1.address", "b.example.com"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.0.name", domainAName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "domain.1.name", domainBName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config:   reversedConfig,
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withGzip(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName := fmt.Sprintf("gzip-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "0"),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithGzip(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.name", gzipName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.0", "text/html"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.1", "text/css"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.0", "css"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.1", "js"),
					// Adding a gzip config should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "0"),
					// Removing the gzip config should create and activate version 3
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_multipleGzips(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName1 := fmt.Sprintf("gzip-a-%s", acctest.RandString(10))
	gzipName2 := fmt.Sprintf("gzip-b-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleGzips(serviceName, domainName, gzipName1, gzipName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.name", gzipName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.1.name", gzipName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.1.extensions.0", "js"),
				),
			},
			{
				Config:   ConfigCDNAutoWithMultipleGzips(serviceName, domainName, gzipName1, gzipName2),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_gzipEmptyLists(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName := fmt.Sprintf("gzip-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				// Explicit empty lists must be preserved rather than normalized to
				// null, otherwise Terraform Core reports "produced inconsistent
				// result after apply" against the planned value.
				Config: ConfigCDNAutoWithGzipEmptyLists(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.name", gzipName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.#", "0"),
				),
			},
			{
				Config:   ConfigCDNAutoWithGzipEmptyLists(serviceName, domainName, gzipName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_gzipRemovedValuesCleared(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName := fmt.Sprintf("gzip-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				// Step 1: populate both content_types and extensions.
				Config: ConfigCDNAutoWithGzip(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.#", "2"),
				),
			},
			{
				// Step 2: remove content_types. A previously configured value being
				// removed must actually clear the remote value, not be silently
				// skipped (see servicecdnauto's use of gzip.ReconcileWithPrevious).
				// The state check alone isn't enough proof: the provider's custom
				// Read normalizes an unset field against the plan regardless of
				// what's actually stored remotely, so it would show "0" either way.
				// CheckGzipFieldClearedRemotely queries the Fastly API directly to
				// confirm the value was genuinely cleared, not just hidden in state.
				Config: ConfigCDNAutoWithGzipContentTypesRemoved(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.#", "2"),
					CheckGzipFieldClearedRemotely("fastly_service_cdn_auto.test", gzipName, "content_types", "text/html text/css"),
				),
			},
			{
				// Step 3: remove extensions too, clearing the last configured value.
				Config: ConfigCDNAutoWithGzipAllRemoved(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.content_types.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.extensions.#", "0"),
					CheckGzipFieldClearedRemotely("fastly_service_cdn_auto.test", gzipName, "content_types", "text/html text/css"),
					CheckGzipFieldClearedRemotely("fastly_service_cdn_auto.test", gzipName, "extensions", "css js"),
				),
			},
			{
				// Step 4: confirm no perpetual diff once both fields are unset.
				Config:   ConfigCDNAutoWithGzipAllRemoved(serviceName, domainName, gzipName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withCondition(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithCondition(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "REQUEST"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.statement", `req.url ~ "^/admin"`),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.priority", "10"),
					// Adding a condition should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
				),
			},
			{
				Config: ConfigCDNAutoWithConditionUpdated(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.statement", `req.url ~ "^/private"`),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.priority", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "0"),
					// Removing the condition should create and activate version 4
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "4"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_multipleConditions(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName1 := fmt.Sprintf("condition-a-%s", acctest.RandString(10))
	conditionName2 := fmt.Sprintf("condition-b-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleConditions(serviceName, domainName, conditionName1, conditionName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "REQUEST"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.1.name", conditionName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.1.type", "CACHE"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.1.priority", "20"),
				),
			},
			{
				// Verify plan stability - reapplying the same config should be a no-op
				Config:   ConfigCDNAutoWithMultipleConditions(serviceName, domainName, conditionName1, conditionName2),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_conditionTypeChange(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithCondition(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "REQUEST"),
					CheckConditionTypeInFastly("fastly_service_cdn_auto.test", conditionName, "REQUEST"),
				),
			},
			{
				// The Fastly API doesn't support updating a condition's type via PUT, so the
				// provider must delete and recreate the condition. Confirm the new type is
				// actually reflected remotely, not just in Terraform state.
				Config: ConfigCDNAutoWithConditionTypeChanged(serviceName, domainName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "CACHE"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.statement", "beresp.status == 200"),
					CheckConditionTypeInFastly("fastly_service_cdn_auto.test", conditionName, "CACHE"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_backendWithRequestCondition(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithBackendRequestCondition(serviceName, domainName, backendName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.request_condition", conditionName),
				),
			},
			{
				// Removing both the backend and the condition it references in the same apply
				// must not fail: the condition is reconciled before the backend so referencing
				// it on create works, but that means a stale condition is deleted before the
				// backend that used it is updated/removed in the same pass. Prove that doesn't
				// trip referential-integrity validation on the Fastly API side.
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "0"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_gzipWithCacheCondition(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName := fmt.Sprintf("gzip-%s", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithGzipCacheCondition(serviceName, domainName, gzipName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "CACHE"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.name", gzipName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.cache_condition", conditionName),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withCacheSetting(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName := fmt.Sprintf("cache-setting-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "0"),
					// Initial version should be 1
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithCacheSetting(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.action", "cache"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "3600"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.stale_ttl", "120"),
					// Adding a cache setting should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "0"),
					// Removing the cache setting should create and activate version 3
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_cacheSettingMinimal(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName := fmt.Sprintf("cache-setting-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				// action/ttl/stale_ttl left unset must default to null/0 rather than
				// drifting on every plan.
				Config: ConfigCDNAutoWithCacheSettingMinimal(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.action"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.stale_ttl", "0"),
				),
			},
			{
				Config:   ConfigCDNAutoWithCacheSettingMinimal(serviceName, domainName, cacheSettingName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_cacheSettingUpdated(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName := fmt.Sprintf("cache-setting-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithCacheSetting(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.action", "cache"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "3600"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.stale_ttl", "120"),
				),
			},
			{
				// Same name, different action/ttl/stale_ttl - this is an in-place update,
				// not a delete+recreate, since the name (the reconciler's identity key)
				// hasn't changed.
				Config: ConfigCDNAutoWithCacheSettingUpdated(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.action", "pass"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "7200"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.stale_ttl", "300"),
				),
			},
			{
				// Removing action/ttl/stale_ttl from config must clear them back to
				// null/0 remotely, not just hide the drift in state.
				Config: ConfigCDNAutoWithCacheSettingMinimal(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.action"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.stale_ttl", "0"),
				),
			},
			{
				Config:   ConfigCDNAutoWithCacheSettingMinimal(serviceName, domainName, cacheSettingName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_multipleCacheSettings(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName1 := fmt.Sprintf("cache-setting-a-%s", acctest.RandString(10))
	cacheSettingName2 := fmt.Sprintf("cache-setting-b-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleCacheSettings(serviceName, domainName, cacheSettingName1, cacheSettingName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.ttl", "3600"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.1.name", cacheSettingName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.1.action", "pass"),
				),
			},
			{
				Config:   ConfigCDNAutoWithMultipleCacheSettings(serviceName, domainName, cacheSettingName1, cacheSettingName2),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_cacheSettingWithCacheCondition(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName := fmt.Sprintf("cache-setting-%s", acctest.RandString(10))
	conditionName := fmt.Sprintf("condition-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithCacheSettingCacheCondition(serviceName, domainName, cacheSettingName, conditionName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.name", conditionName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "condition.0.type", "CACHE"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.cache_condition", conditionName),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withRateLimiter(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithRateLimiter(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.action", "response"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.client_key.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.client_key.0", "req.http.Fastly-Client-IP"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.http_methods.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.http_methods.0", "GET"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.http_methods.1", "POST"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.penalty_box_duration", "10"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.rps_limit", "100"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.window_size", "10"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response.content", "Rate limit exceeded"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response.content_type", "text/plain"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response.status", "429"),
					resource.TestCheckResourceAttrSet("fastly_service_cdn_auto.test", "rate_limiter.0.rate_limiter_id"),
					// Adding a rate limiter should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "0"),
					// Removing the rate limiter should create and activate version 3
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterResponseCleared(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiter(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.action", "response"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response.status", "429"),
				),
			},
			{
				// Clearing response (switching action away from "response") can't be done via
				// UpdateERL - the API rejects an explicit empty value - so this goes through a
				// delete+recreate of the rate limiter instead (see needsRecreate in
				// internal/resources/ratelimiter).
				Config: ConfigCDNAutoWithRateLimiterResponseCleared(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.action", "log_only"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response"),
				),
			},
			{
				Config:   ConfigCDNAutoWithRateLimiterResponseCleared(serviceName, domainName, rateLimiterName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterMinimal(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				// feature_revision left unset must default to 1 rather than drifting on
				// every plan; response/response_object_name/uri_dictionary_name left unset
				// must default to null. logger_type is set here since the Fastly API requires
				// it whenever action is log_only.
				Config: ConfigCDNAutoWithRateLimiterMinimal(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.feature_revision", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.logger_type", "s3"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.response_object_name"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name"),
				),
			},
			{
				Config:   ConfigCDNAutoWithRateLimiterMinimal(serviceName, domainName, rateLimiterName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterUpdated(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiterMinimal(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.penalty_box_duration", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.rps_limit", "50"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.window_size", "60"),
				),
			},
			{
				// Same name, different client_key/http_methods/penalty_box_duration/rps_limit/
				// window_size - this is an in-place update, not a delete+recreate, since the
				// name (the reconciler's identity key) hasn't changed.
				Config: ConfigCDNAutoWithRateLimiterUpdated(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.client_key.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.http_methods.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.penalty_box_duration", "15"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.rps_limit", "75"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.window_size", "1"),
				),
			},
			{
				Config:   ConfigCDNAutoWithRateLimiterUpdated(serviceName, domainName, rateLimiterName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_multipleRateLimiters(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName1 := fmt.Sprintf("rate-limiter-a-%s", acctest.RandString(10))
	rateLimiterName2 := fmt.Sprintf("rate-limiter-b-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithMultipleRateLimiters(serviceName, domainName, rateLimiterName1, rateLimiterName2),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName1),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.action", "log_only"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.1.name", rateLimiterName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.1.action", "response"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.1.response.status", "429"),
				),
			},
			{
				Config:   ConfigCDNAutoWithMultipleRateLimiters(serviceName, domainName, rateLimiterName1, rateLimiterName2),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_withDirector(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	directorName := fmt.Sprintf("director-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "1"),
				),
			},
			{
				Config: ConfigCDNAutoWithDirector(serviceName, domainName, backendName, directorName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.name", directorName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.backends.#", "1"),
					resource.TestCheckTypeSetElemAttr("fastly_service_cdn_auto.test", "director.0.backends.*", backendName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.comment", ""),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.quorum", "75"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.retries", "5"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.shield", ""),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.type", "random"),
					// Adding a director should create and activate version 2
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				// Same name, different comment/quorum/retries/shield/type - this is an in-place
				// update, not a delete+recreate, since the name (the reconciler's identity key)
				// hasn't changed.
				Config: ConfigCDNAutoWithDirectorUpdated(serviceName, domainName, backendName, directorName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.comment", "updated director"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.quorum", "30"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.retries", "10"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.shield", "sjc-ca-us"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.type", "hash"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "0"),
					// Removing the director should create and activate version 4
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "4"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "4"),
				),
			},
		},
	})
}

// TestAccFastlyServiceCDNAuto_directorTypeResetOnOmit is a regression test for dropping an
// explicit director type from config: it must reset to the default ("random"), matching the
// legacy SDKv2 provider on main where type had a schema-level `Default: 1`. It also exercises
// that reset alongside a reorder/insert, since the director block's type plan modifier must
// still resolve each director's prior value by name rather than list position - see the
// typeStickyDefault doc comment in internal/resources/director/schema.go. (The one case where a
// prior value is preserved instead of reset - an existing round_robin director - can't be
// exercised here, since config can never set type = round_robin; that case is covered by the
// name-vs-position unit test for typeStickyDefault in internal/resources/director/director_test.go.)
func TestAccFastlyServiceCDNAuto_directorTypeResetOnOmit(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendNameA := fmt.Sprintf("backend-a-%s", acctest.RandString(10))
	backendNameB := fmt.Sprintf("backend-b-%s", acctest.RandString(10))
	backendNameC := fmt.Sprintf("backend-c-%s", acctest.RandString(10))
	directorNameA := fmt.Sprintf("director-a-%s", acctest.RandString(10))
	directorNameB := fmt.Sprintf("director-b-%s", acctest.RandString(10))
	directorNameC := fmt.Sprintf("director-c-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithTwoOrderedDirectors(serviceName, domainName, backendNameA, backendNameB, directorNameA, directorNameB),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.name", directorNameA),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.type", "hash"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.1.name", directorNameB),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.1.type", "random"),
				),
			},
			{
				// directorC is inserted ahead of directorA, and directorA's explicit type =
				// "hash" is dropped from config. directorA must now reset to "random" - the
				// schema default - rather than sticking with its prior "hash".
				Config: ConfigCDNAutoWithDirectorInsertedAhead(serviceName, domainName, backendNameA, backendNameB, backendNameC, directorNameA, directorNameB, directorNameC),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.name", directorNameC),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.type", "random"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.1.name", directorNameA),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.1.type", "random"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.2.name", directorNameB),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.2.type", "random"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_directorNegativeRetries(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	directorName := fmt.Sprintf("director-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      ConfigCDNAutoWithDirectorNegativeRetries(serviceName, domainName, backendName, directorName),
				ExpectError: regexp.MustCompile(`Attribute director\[0\]\.retries value must be at least 0`),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_directorBackendSwap(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName1 := fmt.Sprintf("backend-a-%s", acctest.RandString(10))
	backendName2 := fmt.Sprintf("backend-b-%s", acctest.RandString(10))
	directorName := fmt.Sprintf("director-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithDirector(serviceName, domainName, backendName1, directorName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.backends.#", "1"),
					resource.TestCheckTypeSetElemAttr("fastly_service_cdn_auto.test", "director.0.backends.*", backendName1),
				),
			},
			{
				// backendName1 is removed from config entirely and backendName2 takes its place
				// as the director's only backend. This only succeeds if the director's backend
				// association is updated before backendName1 is deleted - see the ordering
				// comment in servicecdnauto's Update.
				Config: ConfigCDNAutoWithDirectorBackendSwapped(serviceName, domainName, backendName2, directorName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "backend.0.name", backendName2),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.backends.#", "1"),
					resource.TestCheckTypeSetElemAttr("fastly_service_cdn_auto.test", "director.0.backends.*", backendName2),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_importWithDirector(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-%s", acctest.RandString(10))
	directorName := fmt.Sprintf("director-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithDirector(serviceName, domainName, backendName, directorName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "director.0.name", directorName),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterWithDictionary(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiterDictionary(serviceName, domainName, rateLimiterName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name", dictionaryName),
				),
			},
			{
				// Clearing uri_dictionary_name can't be done via UpdateERL - the API rejects an
				// explicit empty value - so this goes through a delete+recreate of the rate
				// limiter instead (see needsRecreate in internal/resources/ratelimiter).
				Config: ConfigCDNAutoWithRateLimiterDictionaryCleared(serviceName, domainName, rateLimiterName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name"),
				),
			},
			{
				Config:   ConfigCDNAutoWithRateLimiterDictionaryCleared(serviceName, domainName, rateLimiterName, dictionaryName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterDictionaryRemoved(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiterDictionary(serviceName, domainName, rateLimiterName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name", dictionaryName),
				),
			},
			{
				// Removing the dictionary block while the rate limiter's uri_dictionary_name is
				// left unchanged, still naming it: ValidateDictionaryReferences rejects this at
				// plan time, since reconcile.Run only reconciles a rate limiter whose own desired
				// fields changed - leaving this stale would otherwise reach the Fastly API as a
				// version-validation failure (the generated VCL references an undefined table)
				// once the dictionary is actually deleted.
				Config:      ConfigCDNAutoWithRateLimiterDictionaryRemoved(serviceName, domainName, rateLimiterName, dictionaryName),
				ExpectError: regexp.MustCompile(`does not match any configured dictionary`),
			},
			{
				// Revert to a valid config so CheckDestroy's cleanup doesn't hit the same
				// plan-time validation error the previous step intentionally triggered.
				Config: ConfigCDNAutoWithRateLimiterDictionary(serviceName, domainName, rateLimiterName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_rateLimiterDictionaryRemovedTogether(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))
	dictionaryName := fmt.Sprintf("dict_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiterDictionary(serviceName, domainName, rateLimiterName, dictionaryName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.0.name", dictionaryName),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name", dictionaryName),
				),
			},
			{
				// Removing the dictionary block and clearing the rate limiter's
				// uri_dictionary_name together, in the same apply: dictionaries are
				// created/updated, then rate limiters are fully reconciled (clearing the stale
				// reference here), then dictionaries no longer desired are deleted - see the
				// three-pass sequence in servicecdnauto's Update. Confirms the delete succeeds
				// once nothing still references the dictionary.
				Config: ConfigCDNAutoWithRateLimiterMinimal(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "dictionary.#", "0"),
					resource.TestCheckNoResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.uri_dictionary_name"),
				),
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_importWithRateLimiter(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("rate-limiter-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithRateLimiter(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "rate_limiter.0.name", rateLimiterName),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_importWithCacheSetting(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	cacheSettingName := fmt.Sprintf("cache-setting-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithCacheSetting(serviceName, domainName, cacheSettingName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "cache_setting.0.name", cacheSettingName),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_import(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}

func TestAccFastlyServiceCDNAuto_importWithGzip(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	gzipName := fmt.Sprintf("gzip-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigCDNAutoWithGzip(serviceName, domainName, gzipName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "gzip.0.name", gzipName),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
		},
	})
}
