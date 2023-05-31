package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccFastlyServiceVCLProductEnablement_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	backendAddress := "httpbin.org"

	config := fmt.Sprintf(`
  resource "fastly_service_vcl" "foo" {
    name = "%s"

    domain {
      name    = "%s"
      comment = "demo"
    }

    backend {
      address = "%s"
      name    = "%s"
      port    = 443
      shield  = "amsterdam-nl"
    }

    product_enablement {
      brotli_compression = true
      domain_inspector   = false
      image_optimizer    = false
      origin_inspector   = false
      websockets         = false
    }

    force_destroy = true
  }
  `, serviceName, domainName, backendAddress, backendName)

	// The following backends are what we expect to exist after all our Terraform
	// configuration settings have been applied. We expect them to correlate to
	// the specific backend definitions in the Terraform configuration.

	b1 := gofastly.Backend{
		Address: backendAddress,
		Name:    backendName,
		Port:    443,
		Shield:  "amsterdam-nl", // required for image_optimizer

		// NOTE: The following are defaults applied by the API.
		BetweenBytesTimeout: 10000,
		ConnectTimeout:      1000,
		FirstByteTimeout:    15000,
		Hostname:            backendAddress,
		MaxConn:             200,
		SSLCheckCert:        true,
		Weight:              100,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "1"),
					testAccCheckFastlyServiceVCLBackendAttributes(&service, []*gofastly.Backend{&b1}),
				),
			},
		},
	})
}
