package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestAccFastlyServiceProductEnablement_vcl_basic(t *testing.T) {
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
      api_discovery         = false
      bot_management        = false
      brotli_compression    = true
      domain_inspector      = false
      image_optimizer       = false
      log_explorer_insights = false
      origin_inspector      = false
      websockets            = false

      ddos_protection {
        enabled = false
        mode    = "block"
      }

      ngwaf {
        enabled      = false
        workspace_id = "7JFbo4RNA0OKdFWC04r6B3"
        traffic_ramp = 100
      }
    }

    force_destroy = true
  }
  `, serviceName, domainName, backendAddress, backendName)

	// The following backends are what we expect to exist after all our Terraform
	// configuration settings have been applied. We expect them to correlate to
	// the specific backend definitions in the Terraform configuration.

	b1 := gofastly.Backend{
		Address: gofastly.ToPointer(backendAddress),
		Name:    gofastly.ToPointer(backendName),
		Port:    gofastly.ToPointer(443),
		Shield:  gofastly.ToPointer("amsterdam-nl"), // required for image_optimizer

		// NOTE: The following are defaults applied by the API.
		AutoLoadbalance:     gofastly.ToPointer(false),
		BetweenBytesTimeout: gofastly.ToPointer(10000),
		Comment:             gofastly.ToPointer(""),
		ConnectTimeout:      gofastly.ToPointer(1000),
		ErrorThreshold:      gofastly.ToPointer(0),
		FirstByteTimeout:    gofastly.ToPointer(15000),
		HealthCheck:         gofastly.ToPointer(""),
		Hostname:            gofastly.ToPointer(backendAddress),
		MaxConn:             gofastly.ToPointer(200),
		PreferIPv6:          gofastly.ToPointer(false),
		RequestCondition:    gofastly.ToPointer(""),
		SSLCheckCert:        gofastly.ToPointer(true),
		Weight:              gofastly.ToPointer(100),
		UseSSL:              gofastly.ToPointer(false),
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

func TestAccFastlyServiceProductEnablement_vcl_ddosProtectionModeChange(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	backendAddress := "httpbin.org"

	initialConfig := fmt.Sprintf(`
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
    ddos_protection {
      enabled = true
      mode    = "block"
    }
  }

  force_destroy = true
}
`, serviceName, domainName, backendAddress, backendName)

	updatedConfig := fmt.Sprintf(`
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
    ddos_protection {
      enabled = true
      mode    = "log"
    }
  }

  force_destroy = true
}
`, serviceName, domainName, backendAddress, backendName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ddos_protection.0.mode", "block"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ddos_protection.0.mode", "log"),
				),
			},
		},
	})
}

func TestAccFastlyServiceProductEnablement_vcl_ngwafUpdate(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	backendAddress := "httpbin.org"

	initialConfig := fmt.Sprintf(`
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
    ngwaf {
      enabled      = true
      workspace_id = "7JFbo4RNA0OKdFWC04r6B3"
      traffic_ramp = 100
    }
  }

  force_destroy = true
}
`, serviceName, domainName, backendAddress, backendName)

	updatedConfig := fmt.Sprintf(`
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
    ngwaf {
      enabled      = true
      workspace_id = "Jf4Vo9RXd00MdTYJ44xY12"
      traffic_ramp = 80
    }
  }

  force_destroy = true
}
`, serviceName, domainName, backendAddress, backendName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ngwaf.0.workspace_id", "7JFbo4RNA0OKdFWC04r6B3"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ngwaf.0.traffic_ramp", "100"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ngwaf.0.workspace_id", "Jf4Vo9RXd00MdTYJ44xY12"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "product_enablement.0.ngwaf.0.traffic_ramp", "80"),
				),
			},
		},
	})
}

func TestAccFastlyServiceProductEnablement_compute_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	config := fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }

  product_enablement {
    api_discovery         = false
    domain_inspector      = true
    fanout                = false
    log_explorer_insights = false
    websockets            = false

    ddos_protection {
      enabled = false
      mode    = "block"
    }

    ngwaf {
      enabled      = false
      workspace_id = "7JFbo4RNA0OKdFWC04r6B3"
      traffic_ramp = 100
    }
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
  activate = false
}
`, serviceName, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceComputeDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "product_enablement.0.domain_inspector", "true"),
				),
				// Added this flag temporarily until upstream changes
				// are corrected that are causing hash drifts.
				// ref: https://fastly.atlassian.net/browse/CDTOOL-1226
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
