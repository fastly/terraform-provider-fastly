package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccFastlyServiceCDNAuto_imageOptimizerDefaultSettings exercises the full lifecycle of the
// image_optimizer_default_settings nested block on fastly_service_cdn_auto: create, update,
// removal (reset to API defaults), and import. Image Optimizer must be enabled on the service
// before its default settings can be persisted, which in turn requires the test account to be
// allowed to enable Image Optimizer.
func TestAccFastlyServiceCDNAuto_imageOptimizerDefaultSettings(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_cdn_auto"),
		Steps: []resource.TestStep{
			{
				// Bootstrap the service first so Image Optimizer can be enabled on it directly
				// via the API before any image_optimizer_default_settings block is configured.
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					EnableImageOptimizer(t, "fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.#", "0"),
				),
			},
			{
				Config: ConfigCDNAutoWithImageOptimizerDefaultSettings(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.resize_filter", "lanczos3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.webp", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.webp_quality", "85"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.jpeg_type", "auto"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.jpeg_quality", "85"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.upscale", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.allow_video", "false"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "2"),
				),
			},
			{
				ResourceName:            "fastly_service_cdn_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse"},
			},
			{
				Config: ConfigCDNAutoWithImageOptimizerDefaultSettingsUpdated(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.resize_filter", "bicubic"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.webp", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.webp_quality", "70"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.jpeg_type", "progressive"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.jpeg_quality", "90"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.upscale", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.0.allow_video", "true"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "3"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "3"),
				),
			},
			{
				// Removing the block resets Image Optimizer default settings back to API defaults.
				Config: ConfigCDNAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_cdn_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "image_optimizer_default_settings.#", "0"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "active_version", "4"),
					resource.TestCheckResourceAttr("fastly_service_cdn_auto.test", "managed_version", "4"),
				),
			},
		},
	})
}

// TestAccFastlyServiceComputeAuto_imageOptimizerDefaultSettings mirrors
// TestAccFastlyServiceCDNAuto_imageOptimizerDefaultSettings for fastly_service_compute_auto,
// proving the nested block is also supported on the Compute service family.
func TestAccFastlyServiceComputeAuto_imageOptimizerDefaultSettings(t *testing.T) {
	t.Parallel()
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("%s.example.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		CheckDestroy:             CheckServiceDestroy("fastly_service_compute_auto"),
		Steps: []resource.TestStep{
			{
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					EnableImageOptimizer(t, "fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.#", "0"),
				),
			},
			{
				Config: ConfigComputeAutoWithImageOptimizerDefaultSettings(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.resize_filter", "lanczos3"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.webp", "false"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.webp_quality", "85"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.jpeg_type", "auto"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.jpeg_quality", "85"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.upscale", "false"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.allow_video", "false"),
				),
			},
			{
				Config: ConfigComputeAutoWithImageOptimizerDefaultSettingsUpdated(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.resize_filter", "bicubic"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.webp", "true"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.webp_quality", "70"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.jpeg_type", "progressive"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.jpeg_quality", "90"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.upscale", "true"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.0.allow_video", "true"),
				),
			},
			{
				// Removing the block resets Image Optimizer default settings back to API defaults.
				Config: ConfigComputeAutoBasic(serviceName, domainName),
				Check: resource.ComposeTestCheckFunc(
					CheckServiceExists("fastly_service_compute_auto.test"),
					resource.TestCheckResourceAttr("fastly_service_compute_auto.test", "image_optimizer_default_settings.#", "0"),
				),
			},
			{
				ResourceName:            "fastly_service_compute_auto.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "reuse", "package"},
			},
		},
	})
}
