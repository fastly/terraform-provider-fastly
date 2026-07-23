package acceptancetests

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// testNGWAFWorkspaceID mirrors the workspace ID used by the legacy provider's
// own product_enablement acceptance test; it belongs to the Fastly test
// account these acceptance tests already run against.
const testNGWAFWorkspaceID = "7JFbo4RNA0OKdFWC04r6B3"

// TestAccFastlyProductEnablement_cdnBasic creates a CDN service and attaches
// one resource per CDN-applicable product (every product except the
// Compute-only fanout), including the three configurable ones. Each
// resource is independently created, so this also exercises Create for all
// eight simple products plus bot_management, ddos_protection, and ngwaf in
// a single apply.
func TestAccFastlyProductEnablement_cdnBasic(t *testing.T) {
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
				Config: ConfigProductEnablementCDNBasic(serviceName, domainName, backendName, testNGWAFWorkspaceID),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_brotli_compression.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_brotli_compression.test", "id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_image_optimizer.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_origin_inspector.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_domain_inspector.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_websockets.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_log_explorer_insights.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_api_discovery.test", "service_id", "fastly_service_cdn_auto.test", "id"),
					resource.TestCheckResourceAttr("fastly_product_enablement_bot_management.test", "content_guard", "on"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ddos_protection.test", "mode", "block"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "workspace_id", testNGWAFWorkspaceID),
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "traffic_ramp", "50"),
				),
			},
			{
				ResourceName:      "fastly_product_enablement_brotli_compression.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_product_enablement_image_optimizer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_product_enablement_bot_management.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_product_enablement_ddos_protection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_product_enablement_ngwaf.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyProductEnablement_cdnLifecycle verifies the full
// enable/disable lifecycle: starting from a service with nothing enabled,
// adding product resources enables them, and removing the resources again
// disables them - proving that (unlike the single-block predecessor of
// this resource) enablement now tracks resource presence directly, with no
// separate "enabled" attribute.
func TestAccFastlyProductEnablement_cdnLifecycle(t *testing.T) {
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
				Config: ConfigProductEnablementCDNEmpty(serviceName, domainName, backendName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
			{
				Config: ConfigProductEnablementCDNBasic(serviceName, domainName, backendName, testNGWAFWorkspaceID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_brotli_compression.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("fastly_product_enablement_ddos_protection.test", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_product_enablement_ddos_protection.test", "mode", "block"),
				),
			},
			{
				Config: ConfigProductEnablementCDNEmpty(serviceName, domainName, backendName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_brotli_compression.test", plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction("fastly_product_enablement_ddos_protection.test", plancheck.ResourceActionDestroy),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
				),
			},
		},
	})
}

// TestAccFastlyProductEnablement_computeBasic creates a Compute service and
// attaches one resource per Compute-applicable product (every product
// except the CDN-only brotli_compression and image_optimizer).
func TestAccFastlyProductEnablement_computeBasic(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigProductEnablementComputeBasic(serviceName, domainName, testNGWAFWorkspaceID),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_fanout.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_origin_inspector.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_domain_inspector.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_websockets.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_log_explorer_insights.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_api_discovery.test", "service_id", "fastly_service_compute_auto.test", "id"),
					resource.TestCheckResourceAttr("fastly_product_enablement_bot_management.test", "content_guard", "on"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ddos_protection.test", "mode", "block"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "workspace_id", testNGWAFWorkspaceID),
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "traffic_ramp", "100"),
				),
			},
			{
				ResourceName:      "fastly_product_enablement_fanout.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "fastly_product_enablement_ngwaf.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidFanoutOnCDN verifies that
// fastly_product_enablement_fanout is rejected on a CDN service when the
// service and the product resource are created in the same apply.
// service_id is still unknown when ModifyPlan runs in this scenario (the
// service doesn't exist yet), so ModifyPlan defers and the rejection
// instead comes from the same validation running as a fallback in Create,
// after the service itself has already been created. See
// TestAccFastlyProductEnablement_invalidFanoutOnCDNExistingService for the
// counterpart where service_id is already known and ModifyPlan catches it
// during `terraform plan`, before any apply is attempted.
func TestAccFastlyProductEnablement_invalidFanoutOnCDN(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidFanoutOnCDN(serviceName, domainName),
				ExpectError: regexp.MustCompile(`"fanout" is only supported for Compute services`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidFanoutOnCDNExistingService verifies
// that when service_id is already known (the service was created in a
// prior apply), fastly_product_enablement_fanout on a CDN service is
// rejected during `terraform plan` via ModifyPlan, before any apply is
// attempted.
func TestAccFastlyProductEnablement_invalidFanoutOnCDNExistingService(t *testing.T) {
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
				Config:      ConfigProductEnablementInvalidFanoutOnCDNExistingService(serviceName, domainName),
				ExpectError: regexp.MustCompile(`"fanout" is only supported for Compute services`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidBrotliCompressionOnCompute
// verifies that fastly_product_enablement_brotli_compression, a CDN-only
// resource, is rejected on a Compute service - the mirror image of
// TestAccFastlyProductEnablement_invalidFanoutOnCDN.
func TestAccFastlyProductEnablement_invalidBrotliCompressionOnCompute(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidBrotliCompressionOnCompute(serviceName, domainName),
				ExpectError: regexp.MustCompile(`"brotli_compression" is only supported for CDN services`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidImageOptimizerOnCompute verifies
// that fastly_product_enablement_image_optimizer, a CDN-only resource, is
// rejected on a Compute service by ModifyPlan at plan time, before ever
// reaching the live API - which independently rejects it too (confirmed
// manually: "image_optimizer not available for wasm services").
func TestAccFastlyProductEnablement_invalidImageOptimizerOnCompute(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidImageOptimizerOnCompute(serviceName, domainName),
				ExpectError: regexp.MustCompile(`"image_optimizer" is only supported for CDN services`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidNGWAFTrafficRampOnCompute verifies
// that setting traffic_ramp to a non-default value on a Compute service is
// rejected by fastly_product_enablement_ngwaf's ModifyPlan.
func TestAccFastlyProductEnablement_invalidNGWAFTrafficRampOnCompute(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidNGWAFTrafficRampOnCompute(serviceName, domainName, testNGWAFWorkspaceID),
				ExpectError: regexp.MustCompile(`"traffic_ramp" is only supported for CDN services`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidContentGuard verifies that
// fastly_product_enablement_bot_management's content_guard rejects a value
// outside "off"/"on" via the attribute's stringvalidator.OneOf schema
// validation, at plan time and without ever calling the underlying product
// API.
func TestAccFastlyProductEnablement_invalidContentGuard(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidContentGuard(serviceName, domainName),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidDDoSMode verifies that
// fastly_product_enablement_ddos_protection's mode rejects a value outside
// "off"/"log"/"block" via the attribute's stringvalidator.OneOf schema
// validation.
func TestAccFastlyProductEnablement_invalidDDoSMode(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidDDoSMode(serviceName, domainName),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_invalidNGWAFTrafficRampRange verifies that
// fastly_product_enablement_ngwaf's traffic_ramp rejects a value outside
// 0-100 via the attribute's int64validator.Between schema validation.
func TestAccFastlyProductEnablement_invalidNGWAFTrafficRampRange(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config:      ConfigProductEnablementInvalidNGWAFTrafficRampRange(serviceName, domainName, testNGWAFWorkspaceID),
				ExpectError: regexp.MustCompile(`value must be between 0 and 100`),
			},
		},
	})
}

// TestAccFastlyProductEnablement_ddosModeUpdateOnly verifies that changing
// ddos_protection.mode updates fastly_product_enablement_ddos_protection in
// place via UpdateConfiguration, rather than disabling and re-enabling the
// product.
func TestAccFastlyProductEnablement_ddosModeUpdateOnly(t *testing.T) {
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
				Config: ConfigProductEnablementDDoSModeOnly(serviceName, domainName, backendName, "block"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ddos_protection.test", "mode", "block"),
				),
			},
			{
				Config: ConfigProductEnablementDDoSModeOnly(serviceName, domainName, backendName, "log"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_ddos_protection.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_product_enablement_ddos_protection.test", "mode", "log"),
				),
			},
		},
	})
}

// TestAccFastlyProductEnablement_botManagementUpdateContentGuard verifies
// that changing content_guard updates fastly_product_enablement_bot_management
// in place via UpdateConfiguration.
func TestAccFastlyProductEnablement_botManagementUpdateContentGuard(t *testing.T) {
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
				Config: ConfigProductEnablementBotManagementOnly(serviceName, domainName, backendName, "off"),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_product_enablement_bot_management.test", "content_guard", "off"),
				),
			},
			{
				Config: ConfigProductEnablementBotManagementOnly(serviceName, domainName, backendName, "on"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_bot_management.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_product_enablement_bot_management.test", "content_guard", "on"),
				),
			},
		},
	})
}

// TestAccFastlyProductEnablement_ngwafUpdateWorkspaceAndRamp verifies that
// changing traffic_ramp updates fastly_product_enablement_ngwaf in place
// via UpdateConfiguration, without needing to disable and re-enable the
// product or re-link the workspace.
func TestAccFastlyProductEnablement_ngwafUpdateWorkspaceAndRamp(t *testing.T) {
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
				Config: ConfigProductEnablementNGWAFOnly(serviceName, domainName, backendName, testNGWAFWorkspaceID, 25),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "traffic_ramp", "25"),
				),
			},
			{
				Config: ConfigProductEnablementNGWAFOnly(serviceName, domainName, backendName, testNGWAFWorkspaceID, 75),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_ngwaf.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("fastly_product_enablement_ngwaf.test", "traffic_ramp", "75"),
				),
			},
		},
	})
}

// TestAccFastlyProductEnablement_serviceIDForcesReplace verifies that
// changing service_id forces replacement of the resource (destroy against
// the old service, create against the new one) rather than an in-place
// update, per the service_id attribute's RequiresReplace plan modifier.
// Exercised against fastly_product_enablement_domain_inspector as a
// representative simple product.
func TestAccFastlyProductEnablement_serviceIDForcesReplace(t *testing.T) {
	t.Parallel()
	serviceName1 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	serviceName2 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("%s.example.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigProductEnablementServiceIDReplace(serviceName1, domainName1, serviceName2, domainName2, false),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.first"),
					CheckServiceExists("fastly_service_cdn_auto.second"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_domain_inspector.test", "service_id", "fastly_service_cdn_auto.first", "id"),
				),
			},
			{
				Config: ConfigProductEnablementServiceIDReplace(serviceName1, domainName1, serviceName2, domainName2, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fastly_product_enablement_domain_inspector.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.second"),
					resource.TestCheckResourceAttrPair("fastly_product_enablement_domain_inspector.test", "service_id", "fastly_service_cdn_auto.second", "id"),
				),
			},
		},
	})
}
