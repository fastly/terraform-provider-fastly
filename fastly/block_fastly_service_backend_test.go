package fastly

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

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
					Name:                gofastly.ToPointer("test.notexample.com"),
					Address:             gofastly.ToPointer("www.notexample.com"),
					OverrideHost:        gofastly.ToPointer("origin.example.com"),
					Port:                gofastly.ToPointer(80),
					AutoLoadbalance:     gofastly.ToPointer(false),
					BetweenBytesTimeout: gofastly.ToPointer(10000),
					ConnectTimeout:      gofastly.ToPointer(1000),
					ErrorThreshold:      gofastly.ToPointer(0),
					FirstByteTimeout:    gofastly.ToPointer(15000),
					KeepAliveTime:       gofastly.ToPointer(1500),
					MaxConn:             gofastly.ToPointer(200),
					RequestCondition:    gofastly.ToPointer(""),
					HealthCheck:         gofastly.ToPointer(""),
					UseSSL:              gofastly.ToPointer(false),
					ShareKey:            gofastly.ToPointer("sharedkey"),
					SSLCheckCert:        gofastly.ToPointer(true),
					SSLCACert:           gofastly.ToPointer(""),
					SSLCertHostname:     gofastly.ToPointer(""),
					SSLSNIHostname:      gofastly.ToPointer(""),
					SSLClientKey:        gofastly.ToPointer(""),
					SSLClientCert:       gofastly.ToPointer(""),
					MaxTLSVersion:       gofastly.ToPointer(""),
					MinTLSVersion:       gofastly.ToPointer(""),
					SSLCiphers:          gofastly.ToPointer("foo:bar:baz"),
					Shield:              gofastly.ToPointer("lga-ny-us"),
					Weight:              gofastly.ToPointer(100),
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
					"keepalive_time":        1500,
					"max_conn":              200,
					"request_condition":     "",
					"healthcheck":           "",
					"use_ssl":               false,
					"ssl_check_cert":        true,
					"ssl_ca_cert":           "",
					"ssl_cert_hostname":     "",
					"ssl_sni_hostname":      "",
					"ssl_client_key":        "",
					"ssl_client_cert":       "",
					"max_tls_version":       "",
					"min_tls_version":       "",
					"ssl_ciphers":           "foo:bar:baz",
					"share_key":             "sharedkey",
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

func TestResourceFastlyFlattenBackendCompute(t *testing.T) {
	cases := []struct {
		serviceMetadata ServiceMetadata
		remote          []*gofastly.Backend
		local           []map[string]any
	}{
		{
			serviceMetadata: ServiceMetadata{
				serviceType: ServiceTypeCompute,
			},
			remote: []*gofastly.Backend{
				{
					Name:                gofastly.ToPointer("test.notexample.com"),
					Address:             gofastly.ToPointer("www.notexample.com"),
					OverrideHost:        gofastly.ToPointer("origin.example.com"),
					Port:                gofastly.ToPointer(80),
					BetweenBytesTimeout: gofastly.ToPointer(10000),
					ConnectTimeout:      gofastly.ToPointer(1000),
					ErrorThreshold:      gofastly.ToPointer(0),
					FirstByteTimeout:    gofastly.ToPointer(15000),
					KeepAliveTime:       gofastly.ToPointer(1500),
					MaxConn:             gofastly.ToPointer(200),
					HealthCheck:         gofastly.ToPointer(""),
					UseSSL:              gofastly.ToPointer(false),
					SSLCheckCert:        gofastly.ToPointer(true),
					SSLCACert:           gofastly.ToPointer(""),
					SSLCertHostname:     gofastly.ToPointer(""),
					SSLSNIHostname:      gofastly.ToPointer(""),
					SSLClientKey:        gofastly.ToPointer(""),
					SSLClientCert:       gofastly.ToPointer(""),
					MaxTLSVersion:       gofastly.ToPointer(""),
					MinTLSVersion:       gofastly.ToPointer(""),
					SSLCiphers:          gofastly.ToPointer("foo:bar:baz"),
					Shield:              gofastly.ToPointer("lga-ny-us"),
					Weight:              gofastly.ToPointer(100),
				},
			},
			local: []map[string]any{
				{
					"name":                  "test.notexample.com",
					"address":               "www.notexample.com",
					"override_host":         "origin.example.com",
					"port":                  80,
					"between_bytes_timeout": 10000,
					"connect_timeout":       1000,
					"error_threshold":       0,
					"first_byte_timeout":    15000,
					"keepalive_time":        1500,
					"max_conn":              200,
					"healthcheck":           "",
					"use_ssl":               false,
					"ssl_check_cert":        true,
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

func TestAccFastlyServiceVCLBackend_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
	backendAddress := "httpbin.org"

	// The following backends are what we expect to exist after all our Terraform
	// configuration settings have been applied. We expect them to correlate to
	// the specific backend definitions in the Terraform configuration.

	b1 := gofastly.Backend{
		Address:  gofastly.ToPointer(backendAddress),
		Name:     gofastly.ToPointer(backendName),
		Port:     gofastly.ToPointer(443),
		ShareKey: gofastly.ToPointer("sharedkey"),

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
		Shield:              gofastly.ToPointer(""),
		SSLCheckCert:        gofastly.ToPointer(true),
		Weight:              gofastly.ToPointer(100),
		UseSSL:              gofastly.ToPointer(false),
	}
	// This validates the ShareKey is unset.
	b1Updated := gofastly.Backend{
		Address: gofastly.ToPointer(backendAddress),
		Name:    gofastly.ToPointer(backendName),
		Port:    gofastly.ToPointer(443),

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
		Shield:              gofastly.ToPointer(""),
		SSLCheckCert:        gofastly.ToPointer(true),
		Weight:              gofastly.ToPointer(100),
		UseSSL:              gofastly.ToPointer(false),
	}
	b2 := gofastly.Backend{
		Address: gofastly.ToPointer(backendAddress),
		Name:    gofastly.ToPointer(backendName + " new"),
		Port:    gofastly.ToPointer(443),

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
		Shield:              gofastly.ToPointer(""),
		SSLCheckCert:        gofastly.ToPointer(true),
		Weight:              gofastly.ToPointer(100),
		UseSSL:              gofastly.ToPointer(false),
	}
	b3 := gofastly.Backend{
		Address: gofastly.ToPointer(backendAddress),
		Name:    gofastly.ToPointer(backendName + " new with use ssl"),
		// NOTE: We don't set the port attribute in the Terraform configuration, and
		// so the Terraform provider defaults to setting that to port 80. This test
		// validates that the Fastly API currently accepts port 80 (although the
		// setting of use_ssl would otherwise cause you to expect some kind of API
		// validation to prevent port 80 from being used).
		Port:            gofastly.ToPointer(80),
		SSLCertHostname: gofastly.ToPointer("httpbin.org"),
		UseSSL:          gofastly.ToPointer(true),

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
		Shield:              gofastly.ToPointer(""),
		SSLCheckCert:        gofastly.ToPointer(true),
		Weight:              gofastly.ToPointer(100),
	}
	b4 := gofastly.Backend{
		Address:    gofastly.ToPointer(backendAddress),
		Name:       gofastly.ToPointer(backendName + " ipv6"),
		Port:       gofastly.ToPointer(443),
		PreferIPv6: gofastly.ToPointer(true),

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
		RequestCondition:    gofastly.ToPointer(""),
		Shield:              gofastly.ToPointer(""),
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
				Config: testAccServiceVCLBackend(serviceName, domainName, backendAddress, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "1"),
					testAccCheckFastlyServiceVCLBackendAttributes(&service, []*gofastly.Backend{&b1}),
				),
			},

			{
				Config: testAccServiceVCLBackendUpdate(serviceName, domainName, backendAddress, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "4"),
					testAccCheckFastlyServiceVCLBackendAttributes(&service, []*gofastly.Backend{&b1Updated, &b2, &b3, &b4}),
				),
			},
		},
	})
}

// NOTE: We set the port to 443 so we can validate the API is expecting SSL/TLS
// related attributes to not be accidentally sent with empty strings.
func testAccServiceVCLBackend(serviceName, domainName, backendAddress, backendName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address   = "%s"
    name      = "%s"
    port      = 443
    share_key = "sharedkey"
  }

  force_destroy = true
}`, serviceName, domainName, backendAddress, backendName)
}

// NOTE: We set the port to 443 so we can validate the API is expecting SSL/TLS
// related attributes to not be accidentally sent with empty strings.
func testAccServiceVCLBackendUpdate(serviceName, domainName, backendAddress, backendName string) string {
	return fmt.Sprintf(`
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
  }

  backend {
    address = "%s"
    name    = "%s new"
    port    = 443
  }

  backend {
    address           = "%s"
    name              = "%s new with use ssl"
    use_ssl           = true
    ssl_cert_hostname = "httpbin.org"
  }

  backend {
    address     = "%s"
    name        = "%s ipv6"
    port        = 443
    prefer_ipv6 = true
  }

  force_destroy = true
}`, serviceName, domainName, backendAddress, backendName, backendAddress, backendName, backendAddress, backendName, backendAddress, backendName)
}

func testAccCheckFastlyServiceVCLBackendAttributes(service *gofastly.ServiceDetail, want []*gofastly.Backend) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		have, err := conn.ListBackends(&gofastly.ListBackendsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(have) != len(want) {
			return fmt.Errorf("backend list count mismatch, expected (%d), got (%d)", len(want), len(have))
		}

		var found int
		for _, w := range want {
			for _, h := range have {
				if gofastly.ToValue(w.Name) == gofastly.ToValue(h.Name) {
					// we don't know these things ahead of time, so populate them now
					w.ServiceID = service.ServiceID
					w.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					h.CreatedAt = nil
					h.UpdatedAt = nil
					if !reflect.DeepEqual(w, h) {
						return fmt.Errorf("bad match Backend match, expected (%#v), got (%#v)", w, h)
					}
					found++
				}
			}
		}

		if found != len(want) {
			return fmt.Errorf("error matching Backends (%d/%d)", found, len(want))
		}

		return nil
	}
}
