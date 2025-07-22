package fastly

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestAccFastlyServiceImageOptimizerDefaultSettings_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	backendAddress := "httpbin.org"

	block1 := `
		resize_filter = "lanczos2"
		webp = true
		webp_quality = 100
		jpeg_type = "progressive"
		jpeg_quality = 100
		upscale = true
		allow_video = false
	`

	defaultSettings1 := gofastly.ImageOptimizerDefaultSettings{
		ResizeFilter: "lanczos2",
		Webp:         true,
		WebpQuality:  100,
		JpegType:     "progressive",
		JpegQuality:  100,
		Upscale:      true,
		AllowVideo:   false,
	}

	block2 := `
		resize_filter = "bicubic"
		webp = false
		webp_quality = 30
		jpeg_type = "baseline"
		jpeg_quality = 20
		upscale = true
		allow_video = true
	`

	defSettings2 := gofastly.ImageOptimizerDefaultSettings{
		ResizeFilter: "bicubic",
		Webp:         false,
		WebpQuality:  30,
		JpegType:     "baseline",
		JpegQuality:  20,
		Upscale:      true,
		AllowVideo:   true,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageOptimizerDefaultSettingsVCLConfig(serviceName, domainName, backendAddress, backendName, block1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "image_optimizer_default_settings.#", "1"),
					testAccCheckFastlyServiceImageOptimizerDefaultSettingsAttributes(&service, &defaultSettings1),
				),
			},

			{
				Config: testAccImageOptimizerDefaultSettingsVCLConfig(serviceName, domainName, backendAddress, backendName, block2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "image_optimizer_default_settings.#", "1"),
					testAccCheckFastlyServiceImageOptimizerDefaultSettingsAttributes(&service, &defSettings2),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceImageOptimizerDefaultSettingsAttributes(service *gofastly.ServiceDetail, want *gofastly.ImageOptimizerDefaultSettings) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		have, err := conn.GetImageOptimizerDefaultSettings(context.TODO(), &gofastly.GetImageOptimizerDefaultSettingsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Image Optimizer default settings for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if !reflect.DeepEqual(want, have) {
			return fmt.Errorf("bad Image Optimizer default settings, expected (%#v), got (%#v)", want, have)
		}

		return nil
	}
}

func testAccImageOptimizerDefaultSettingsVCLConfig(serviceName, domainName, backendAddress, backendName, imageOptimizerSettings string) string {
	return fmt.Sprintf(`
	resource "fastly_service_vcl" "foo" {
	  name = "%s"

	  domain {
		name	= "%s"
		comment = "demo"
	  }

	  backend {
		address = "%s"
		name	= "%s"
		port	= 443
		shield  = "amsterdam-nl"
	  }

	  image_optimizer_default_settings {
		%s
	  }

	  product_enablement {
		image_optimizer = true
	  }

	  force_destroy = true
	}`, serviceName, domainName, backendAddress, backendName, imageOptimizerSettings)
}
