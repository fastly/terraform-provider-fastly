package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("fastly_service_vcl", &resource.Sweeper{
		Name: "fastly_service_vcl",
		F:    testSweepServices,
	})
}

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    "test.notexample.com",
					Comment: "not comment",
				},
			},
			local: []map[string]any{
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
			local: []map[string]any{
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
		local           []map[string]any
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
					AutoLoadbalance:     false,
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
					SSLCiphers:          "foo:bar:baz",
					Shield:              "lga-ny-us",
					Weight:              uint(100),
				},
			},
			local: []map[string]any{
				{
					"name":                  "test.notexample.com",
					"address":               "www.notexample.com",
					"override_host":         "origin.example.com",
					"port":                  80,
					"auto_loadbalance":      false,
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
					"ssl_ciphers":           "foo:bar:baz",
					"shield":                "lga-ny-us",
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

func TestAccFastlyServiceVCL_updateDomain(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	nameUpdate := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName3 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributes(&service, name, []string{domainName1}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "domain.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLConfigDomainAdd(nameUpdate, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributes(&service, nameUpdate, []string{domainName1, domainName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", nameUpdate),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "domain.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLConfigDomainUpdateComment(nameUpdate, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributes(&service, nameUpdate, []string{domainName1, domainName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "3"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "domain.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLConfigDomainUpdate(nameUpdate, domainName1, domainName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributes(&service, nameUpdate, []string{domainName1, domainName3}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "4"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "domain.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_updateBackend(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBackend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName}),
				),
			},

			{
				Config: testAccServiceVCLConfigBackendUpdate(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "backend.#", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceVCL_activateNewVersionExternally tests whether things break when a new version is cloned and
// activated outside of Terraform. There has been a bug where the version used for reading the state, and the version
// that gets cloned in order to make updates, are different when a new version is activated externally. In this case, a
// 409 conflict error is produced by this test because it reads the new version when making a plan, plans to re-add in
// the deleted backend, but clones the original version which still had the backend and fails with a conflict.
func TestAccFastlyServiceVCL_activateNewVersionExternally(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	activateNewVersion := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		version, err := conn.CloneVersion(&gofastly.CloneVersionInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return err
		}

		err = conn.DeleteBackend(&gofastly.DeleteBackendInput{
			ServiceID:      service.ID,
			ServiceVersion: version.Number,
			Name:           "tf-test-backend",
		})
		if err != nil {
			return err
		}

		_, err = conn.ActivateVersion(&gofastly.ActivateVersionInput{
			ServiceID:      service.ID,
			ServiceVersion: version.Number,
		})
		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBackendUpdate(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName, backendName2}),
					activateNewVersion,
				),
				ExpectNonEmptyPlan: true,
			},

			{
				Config: testAccServiceVCLConfigBackendUpdate(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "3"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "backend.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_updateInvalidBackend(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	badBackendName := fmt.Sprintf("%s.aws.amazon.com.", acctest.RandString(3))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceVCLConfigBackend(name, domain, badBackendName),
				ExpectError: regexp.MustCompile("Bad Request"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{}),
				),
			},

			{
				Config: testAccServiceVCLConfigBackendUpdate(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "backend.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_createServiceWithStaticBackend(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	snippet1 := `
	backend F_httpbin_org {
		.always_use_host_header = true;
		.between_bytes_timeout = 1s;
		.connect_timeout = 1s;
		.dynamic = true;
		.first_byte_timeout = 1s;
		.host = "httpbin.org";
		.host_header = "httpbin.org";
		.max_connections = 200;
		.port = "443";
		.share_key = "foo";
		.ssl = true;
		.ssl_cert_hostname = "httpbin.org";
		.ssl_check_cert = always;
		.ssl_sni_hostname = "httpbin.org";
		.probe = {
			.dummy = true;
			.initial = 5;
			.request = "HEAD / HTTP/1.1"  "Host: httpbin.org" "Connection: close";
			.threshold = 1;
			.timeout = 2s;
			.window = 5;
		  }
	}
	`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigStaticBackend(name, domain, snippet1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "backend.#", "0"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	comment := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	versionComment1 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	versionComment2 := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigInitWithVersionComment(name, versionComment1, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "version_comment", versionComment1),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "1"),
				),
			},
			{
				// updating service comment with "activate = false" will be ignored
				Config: testAccServiceVCLConfigUpdateServiceComment(name, "new service comment", domainName1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "active_version", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "comment", "Managed by Terraform"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccServiceVCLConfigBasicUpdate(name, comment, versionComment2, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "comment", comment),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "version_comment", versionComment2),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "active_version", "2"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "1"),
				),
			},
			{
				ResourceName:      "fastly_service_vcl.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "imported"},
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return fmt.Sprintf("%s@2", service.ID), nil
				},
			},
		},
	})
}

// ServiceVCL_disappears – test that a non-empty plan is returned when a Fastly
// Service is destroyed outside of Terraform, and can no longer be found,
// correctly clearing the ID field and generating a new plan
func TestAccFastlyServiceVCL_disappears(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the service
		conn := testAccProvider.Meta().(*APIClient).conn
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
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfig(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceVCLExists(n string, service *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Service ID is set")
		}

		conn := testAccProvider.Meta().(*APIClient).conn
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

func testAccCheckFastlyServiceVCLAttributes(service *gofastly.ServiceDetail, name string, domains []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Domains for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
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
			return fmt.Errorf("domain count mismatch, expected: %#v, got: %#v", domains, domainList)
		}

		return nil
	}
}

func testAccCheckFastlyServiceVCLAttributesBackends(service *gofastly.ServiceDetail, name string, backends []string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if service.Name != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
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
			return fmt.Errorf("backend count mismatch, expected: %#v, got: %#v", backends, backendList)
		}

		return nil
	}
}

func TestAccFastlyServiceVCL_defaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			// service default is 3600
			{
				Config: testAccServiceVCLConfigBackend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_ttl", "3600"),
				),
			},
			// update default TTL
			{
				Config: testAccServiceVCLConfigBackendTTL(name, domain, backendName, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_ttl", "3400"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "2"),
				),
			},
			// can set 0
			{
				Config: testAccServiceVCLConfigBackendTTL(name, domain, backendName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_ttl", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_defaultHost(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	defaultHost := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigDefaultHost(name, domain, defaultHost),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_ttl", "3600"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "stale_if_error_ttl", "43200"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_host", defaultHost),
				),
			},
			// remove default_host
			{
				Config: testAccServiceVCLConfigDefaultHost(name, domain, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_host", ""),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "active_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceVCL_brokenSnippet tests that a service can still be updated after it has failed during an apply.
// This avoids a bug when activate=true, where setting an invalid snippet causes the resourceServiceUpdate function to
// return early before activating the version. This broke the assumption that cloned_version always tracks the active
// version when activate=true, and means that the version we read from, and the one we clone from in order to make changes,
// are different, meaning the plan is applied to a different version and 409 conflict errors can occur.
func TestAccFastlyServiceVCL_brokenSnippet(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.test", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBrokenSnippet(name, domain, "backend1", `if (req.url !~ "^/anything") {
                       set req.url = "/anything" req.url;
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
				),
			},
			{
				Config: testAccServiceVCLConfigBrokenSnippet(name, domain, "backend2", `if (req.url !~ "^/anything") {
                       set req.url = "/anything" req.url
                     }`),
				ExpectError: regexp.MustCompile(`invalid configuration for Fastly Service`),
			},
			{
				Config: testAccServiceVCLConfigBrokenSnippet(name, domain, "backend2", `if (req.url !~ "^/anything") {
                       set req.url = "/anything" req.url;
                     }`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_createZeroDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLConfigBackendZeroTTL(name, domain, backendName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLAttributesBackends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "default_ttl", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "stale_if_error", "true"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "stale_if_error_ttl", "0"),
				),
			},
		},
	})
}

func testAccCheckServiceVCLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_vcl" {
			continue
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("error listing services when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccServiceVCLConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigUpdateServiceComment(name, comment string, domain string, activate bool) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  comment = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  activate = %t
  force_destroy = true
}`, name, comment, domain, activate)
}

func testAccServiceVCLConfigInitWithVersionComment(name, versionComment, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
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
}`, name, versionComment, domain)
}

func testAccServiceVCLConfigDefaultHost(name, domain, defaultHost string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  default_host = "%s"

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain, defaultHost)
}

func testAccServiceVCLConfigBasicUpdate(name, comment, versionComment, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigDomainAdd(name, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigDomainUpdateComment(name, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigDomainUpdate(name, domain1, domain3 string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigBackend(name, domain, backend string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

func testAccServiceVCLConfigStaticBackend(name, domain, snippet string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
    name           = "%s"
    force_destroy = true

    domain {
        name = "%s"
    }

    snippet {
        content  = <<-EOT
            %s
        EOT
        name     = "vcl_init"
        priority = 50
        type     = "init"
    }
}`, name, domain, snippet)
}

func testAccServiceVCLConfigBackendTTL(name, domain, backend string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  default_ttl = %d
  stale_if_error = true

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

func testAccServiceVCLConfigBackendZeroTTL(name, domain, backend string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"
  default_ttl = %d
  stale_if_error = true
  stale_if_error_ttl = 0

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

func testAccServiceVCLConfigBackendUpdate(name, domain, backend, backend2 string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

	default_ttl = %d
	stale_if_error = true

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

func testAccServiceVCLConfigBrokenSnippet(name, domain, backendName, snippet string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
    name           = "%s"
    activate       = true
    force_destroy = true

    backend {
        address = "httpbin.org"
        name = "%s"
    }

    domain {
        name = "%s"
    }

    snippet {
        content  = <<-EOT
            %s
        EOT
        name     = "url rewrite"
        priority = 100
        type     = "recv"
    }
}`, name, backendName, domain, snippet)
}

func testSweepServices(region string) error {
	client, diagnostics := sharedClientForRegion(region)
	if diagnostics.HasError() {
		return diagToErr(diagnostics)
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
