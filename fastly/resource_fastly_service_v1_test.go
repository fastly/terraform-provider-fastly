package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("fastly_service_v1", &resource.Sweeper{
		Name: "fastly_service_v1",
		F:    testSweepServices,
	})
}

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    "test.notexample.com",
					Comment: "not comment",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "not comment",
				},
			},
		},
		{
			remote: []*gofastly.Domain{
				{
					Name: "test.notexample.com",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDomains(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestResourceFastlyFlattenBackend(t *testing.T) {
	cases := []struct {
		serviceMetadata ServiceMetadata
		remote          []*gofastly.Backend
		local           []map[string]interface{}
	}{
		{
			serviceMetadata: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: []*gofastly.Backend{
				{
					Name:                "test.notexample.com",
					Address:             "www.notexample.com",
					OverrideHost:        "origin.example.com",
					Port:                uint(80),
					AutoLoadbalance:     true,
					BetweenBytesTimeout: uint(10000),
					ConnectTimeout:      uint(1000),
					ErrorThreshold:      uint(0),
					FirstByteTimeout:    uint(15000),
					MaxConn:             uint(200),
					RequestCondition:    "",
					HealthCheck:         "",
					UseSSL:              false,
					SSLCheckCert:        true,
					SSLHostname:         "",
					SSLCACert:           "",
					SSLCertHostname:     "",
					SSLSNIHostname:      "",
					SSLClientKey:        "",
					SSLClientCert:       "",
					MaxTLSVersion:       "",
					MinTLSVersion:       "",
					SSLCiphers:          []string{"foo", "bar", "baz"},
					Shield:              "New York",
					Weight:              uint(100),
				},
			},
			local: []map[string]interface{}{
				{
					"name":                  "test.notexample.com",
					"address":               "www.notexample.com",
					"override_host":         "origin.example.com",
					"port":                  80,
					"auto_loadbalance":      true,
					"between_bytes_timeout": 10000,
					"connect_timeout":       1000,
					"error_threshold":       0,
					"first_byte_timeout":    15000,
					"max_conn":              200,
					"request_condition":     "",
					"healthcheck":           "",
					"use_ssl":               false,
					"ssl_check_cert":        true,
					"ssl_hostname":          "",
					"ssl_ca_cert":           "",
					"ssl_cert_hostname":     "",
					"ssl_sni_hostname":      "",
					"ssl_client_key":        "",
					"ssl_client_cert":       "",
					"max_tls_version":       "",
					"min_tls_version":       "",
					"ssl_ciphers":           "foo,bar,baz",
					"shield":                "New York",
					"weight":                100,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBackend(c.remote, c.serviceMetadata)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_updateDomain(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	nameUpdate := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName3 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, name, []string{domainName1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "1"),
				),
			},

			{
				Config: testAccServiceV1Config_domainAdd(nameUpdate, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, nameUpdate, []string{domainName1, domainName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", nameUpdate),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "2"),
				),
			},

			{
				Config: testAccServiceV1Config_domainUpdateComment(nameUpdate, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, nameUpdate, []string{domainName1, domainName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "3"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "2"),
				),
			},

			{
				Config: testAccServiceV1Config_domainUpdate(nameUpdate, domainName1, domainName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, nameUpdate, []string{domainName1, domainName3}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "4"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_updateBackend(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
				),
			},

			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "backend.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_updateInvalidBackend(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	badBackendName := fmt.Sprintf("%s.aws.amazon.com.", acctest.RandString(3))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceV1Config_backend(name, domain, badBackendName),
				ExpectError: regexp.MustCompile("Bad Request"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{}),
				),
			},

			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "backend.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	comment := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	versionComment := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "version_comment", ""),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "backend.#", "1"),
				),
			},

			{
				Config: testAccServiceV1Config_basicUpdate(name, comment, versionComment, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "comment", comment),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "version_comment", versionComment),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "backend.#", "1"),
				),
			},
		},
	})
}

// ServiceV1_disappears – test that a non-empty plan is returned when a Fastly
// Service is destroyed outside of Terraform, and can no longer be found,
// correctly clearing the ID field and generating a new plan
func TestAccFastlyServiceV1_disappears(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the service
		conn := testAccProvider.Meta().(*FastlyClient).conn
		// deactivate active version to destoy
		_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return err
		}

		// delete service
		return conn.DeleteService(&gofastly.DeleteServiceInput{
			ID: service.ID,
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceV1Exists(n string, service *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latest, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*service = *latest

		return nil
	}
}

func testAccCheckFastlyServiceV1Attributes(service *gofastly.ServiceDetail, name string, domains []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expected := len(domains)
		for _, d := range domainList {
			for _, e := range domains {
				if d.Name == e {
					expected--
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Domain count mismatch, expected: %#v, got: %#v", domains, domainList)
		}

		return nil
	}
}

func testAccCheckFastlyServiceV1Attributes_backends(service *gofastly.ServiceDetail, name string, backends []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expected := len(backendList)
		for _, b := range backendList {
			for _, e := range backends {
				if b.Address == e {
					expected--
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Backend count mismatch, expected: %#v, got: %#v", backends, backendList)
		}

		return nil
	}
}

func TestAccFastlyServiceV1_defaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3600"),
				),
			},

			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3400"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
				),
			},
			// Now update the default_ttl to 0 and encounter the issue https://github.com/hashicorp/terraform/issues/12910
			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_createDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backendTTL(name, domain, backendName, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3400"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_createZeroDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backendTTL(name, domain, backendName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "0"),
				),
			},
		},
	})
}

func testAccCheckServiceV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing servcies when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func TestAccFastlyServiceV1_import(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1_import(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
				),
			},
			{
				ResourceName:      "fastly_service_v1.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy"},
			},
		},
	})

}

func testAccServiceV1Config(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceV1Config_basicUpdate(name, comment, versionComment, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name    = "%s"
  comment = "%s"
  version_comment = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, comment, versionComment, domain)
}

func testAccServiceV1Config_domainAdd(name, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  domain {
    name    = "%s"
    comment = "tf-testing-other-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain1, domain2)
}

func testAccServiceV1Config_domainUpdateComment(name, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain-updated"
  }

  domain {
    name    = "%s"
    comment = "tf-testing-other-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain1, domain2)
}

func testAccServiceV1Config_domainUpdate(name, domain1, domain3 string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain-updated"
  }

  domain {
    name    = "%s"
    comment = "tf-testing-other-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain1, domain3)
}

func testAccServiceV1Config_backend(name, domain, backend string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  force_destroy = true
}`, name, domain, backend)
}

func testAccServiceV1Config_backendTTL(name, domain, backend string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  default_ttl = %d

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  force_destroy = true
}`, name, ttl, domain, backend)
}

func testAccServiceV1Config_backend_update(name, domain, backend, backend2 string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

	default_ttl = %d

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test-backend"
  }

  backend {
    address = "%s"
    name    = "tf-test-backend-other"
  }

  force_destroy = true
}`, name, ttl, domain, backend, backend2)
}

func testAccServiceV1_import(serviceName string) string {
	backendName01 := fmt.Sprintf("%s.%s.com", acctest.RandString(3), acctest.RandString(12))
	domainName01 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName02 := fmt.Sprintf("%s.%s.com", acctest.RandString(3), acctest.RandString(12))
	domainName02 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain01"
	}
  domain {
    name    = "%s"
    comment = "tf-testing-domain02"
	}

  backend {
    address = "%s"
    name    = "tf -test backend 01"
  }

  backend {
    address = "%s"
    name    = "tf -test backend 02"
  }

  force_destroy = true
}

`, serviceName, domainName01, backendName01, domainName02, backendName02)
}

func testSweepServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	services, err := client.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return err
	}

	for _, service := range services {
		if strings.HasPrefix(service.Name, testResourcePrefix) {
			s, err := client.GetServiceDetails(&gofastly.GetServiceInput{
				ID: service.ID,
			})
			if err != nil {
				return err
			}

			if s.ActiveVersion.Number != 0 {
				_, err := client.DeactivateVersion(&gofastly.DeactivateVersionInput{
					ServiceID:      service.ID,
					ServiceVersion: s.ActiveVersion.Number,
				})
				if err != nil {
					return err
				}
			}

			err = client.DeleteService(&gofastly.DeleteServiceInput{
				ID: service.ID,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
