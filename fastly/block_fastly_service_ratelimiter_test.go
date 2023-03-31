package fastly

import (
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
)

func TestResourceFastlyFlattenRateLimiter(t *testing.T) {
	cases := []struct {
		serviceMetadata ServiceMetadata
		remote          []*gofastly.ERL
		local           []map[string]any
	}{
		{
			serviceMetadata: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: []*gofastly.ERL{
				{
					Action: gofastly.ERLActionResponse,
					ClientKey: []string{
						"req.http.Fastly-Client-IP",
						"req.http.User-Agent",
					},
					FeatureRevision: 1,
					HTTPMethods: []string{
						"POST",
						"PUT",
						"PATCH",
						"DELETE",
					},
					ID:                 "123abc",
					LoggerType:         gofastly.ERLLogBigQuery,
					Name:               "example",
					PenaltyBoxDuration: 123,
					Response: &gofastly.ERLResponse{
						ERLContent:     "example",
						ERLContentType: "plain/text",
						ERLStatus:      429,
					},
					ResponseObjectName: "example",
					RpsLimit:           123,
					WindowSize:         gofastly.ERLSize1,
				},
			},
			local: []map[string]any{
				{
					"action":               "response",
					"client_key":           "req.http.Fastly-Client-IP,req.http.User-Agent",
					"feature_revision":     1,
					"http_methods":         "POST,PUT,PATCH,DELETE",
					"logger_type":          "bigquery",
					"name":                 "example",
					"penalty_box_duration": 123,
					"ratelimiter_id":       "123abc",
					"response": []map[string]any{
						{
							"content":      "example",
							"content_type": "plain/text",
							"status":       429,
						},
					},
					"response_object_name": "example",
					"rps_limit":            123,
					"window_size":          1,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenRateLimiter(c.remote, c.serviceMetadata)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

//
// func TestResourceFastlyFlattenBackendCompute(t *testing.T) {
// 	cases := []struct {
// 		serviceMetadata ServiceMetadata
// 		remote          []*gofastly.Backend
// 		local           []map[string]any
// 	}{
// 		{
// 			serviceMetadata: ServiceMetadata{
// 				serviceType: ServiceTypeCompute,
// 			},
// 			remote: []*gofastly.Backend{
// 				{
// 					Name:                "test.notexample.com",
// 					Address:             "www.notexample.com",
// 					OverrideHost:        "origin.example.com",
// 					Port:                80,
// 					BetweenBytesTimeout: 10000,
// 					ConnectTimeout:      1000,
// 					ErrorThreshold:      0,
// 					FirstByteTimeout:    15000,
// 					KeepAliveTime:       1500,
// 					MaxConn:             200,
// 					HealthCheck:         "",
// 					UseSSL:              false,
// 					SSLCheckCert:        true,
// 					SSLHostname:         "",
// 					SSLCACert:           "",
// 					SSLCertHostname:     "",
// 					SSLSNIHostname:      "",
// 					SSLClientKey:        "",
// 					SSLClientCert:       "",
// 					MaxTLSVersion:       "",
// 					MinTLSVersion:       "",
// 					SSLCiphers:          "foo:bar:baz",
// 					Shield:              "lga-ny-us",
// 					Weight:              100,
// 				},
// 			},
// 			local: []map[string]any{
// 				{
// 					"name":                  "test.notexample.com",
// 					"address":               "www.notexample.com",
// 					"override_host":         "origin.example.com",
// 					"port":                  80,
// 					"between_bytes_timeout": 10000,
// 					"connect_timeout":       1000,
// 					"error_threshold":       0,
// 					"first_byte_timeout":    15000,
// 					"keepalive_time":        1500,
// 					"max_conn":              200,
// 					"healthcheck":           "",
// 					"use_ssl":               false,
// 					"ssl_check_cert":        true,
// 					"ssl_ca_cert":           "",
// 					"ssl_cert_hostname":     "",
// 					"ssl_sni_hostname":      "",
// 					"ssl_client_key":        "",
// 					"ssl_client_cert":       "",
// 					"max_tls_version":       "",
// 					"min_tls_version":       "",
// 					"ssl_ciphers":           "foo:bar:baz",
// 					"shield":                "lga-ny-us",
// 					"weight":                100,
// 				},
// 			},
// 		},
// 	}
//
// 	for _, c := range cases {
// 		out := flattenBackend(c.remote, c.serviceMetadata)
// 		if !reflect.DeepEqual(out, c.local) {
// 			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
// 		}
// 	}
// }
//
// func TestAccFastlyServiceVCLBackend_basic(t *testing.T) {
// 	var service gofastly.ServiceDetail
// 	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
// 	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
// 	backendName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))
// 	backendAddress := "httpbin.org"
//
// 	// The following backends are what we expect to exist after all our Terraform
// 	// configuration settings have been applied. We expect them to correlate to
// 	// the specific backend definitions in the Terraform configuration.
//
// 	b1 := gofastly.Backend{
// 		Address: backendAddress,
// 		Name:    backendName,
// 		Port:    443,
//
// 		// NOTE: The following are defaults applied by the API.
// 		BetweenBytesTimeout: 10000,
// 		ConnectTimeout:      1000,
// 		FirstByteTimeout:    15000,
// 		Hostname:            backendAddress,
// 		MaxConn:             200,
// 		SSLCheckCert:        true,
// 		Weight:              100,
// 	}
// 	b2 := gofastly.Backend{
// 		Address: backendAddress,
// 		Name:    backendName + " new",
// 		Port:    443,
//
// 		// NOTE: The following are defaults applied by the API.
// 		BetweenBytesTimeout: 10000,
// 		ConnectTimeout:      1000,
// 		FirstByteTimeout:    15000,
// 		Hostname:            backendAddress,
// 		MaxConn:             200,
// 		SSLCheckCert:        true,
// 		Weight:              100,
// 	}
// 	b3 := gofastly.Backend{
// 		Address: backendAddress,
// 		Name:    backendName + " new with use ssl",
// 		// NOTE: We don't set the port attribute in the Terraform configuration, and
// 		// so the Terraform provider defaults to setting that to port 80. This test
// 		// validates that the Fastly API currently accepts port 80 (although the
// 		// setting of use_ssl would otherwise cause you to expect some kind of API
// 		// validation to prevent port 80 from being used).
// 		Port:            80,
// 		SSLCertHostname: "httpbin.org",
// 		UseSSL:          true,
//
// 		// NOTE: The following are defaults applied by the API.
// 		BetweenBytesTimeout: 10000,
// 		ConnectTimeout:      1000,
// 		FirstByteTimeout:    15000,
// 		Hostname:            backendAddress,
// 		MaxConn:             200,
// 		SSLCheckCert:        true,
// 		Weight:              100,
// 	}
//
// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			testAccPreCheck(t)
// 		},
// 		ProviderFactories: testAccProviders,
// 		CheckDestroy:      testAccCheckServiceVCLDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccServiceVCLBackend(serviceName, domainName, backendAddress, backendName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
// 					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", serviceName),
// 					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "1"),
// 					testAccCheckFastlyServiceVCLBackendAttributes(&service, []*gofastly.Backend{&b1}),
// 				),
// 			},
//
// 			{
// 				Config: testAccServiceVCLBackendUpdate(serviceName, domainName, backendAddress, backendName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
// 					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "backend.#", "3"),
// 					testAccCheckFastlyServiceVCLBackendAttributes(&service, []*gofastly.Backend{&b1, &b2, &b3}),
// 				),
// 			},
// 		},
// 	})
// }
//
// // NOTE: We set the port to 443 so we can validate the API is expecting SSL/TLS
// // related attributes to not be accidentally sent with empty strings.
// func testAccServiceVCLBackend(serviceName, domainName, backendAddress, backendName string) string {
// 	return fmt.Sprintf(`
// resource "fastly_service_vcl" "foo" {
//   name = "%s"
//
//   domain {
//     name    = "%s"
//     comment = "demo"
//   }
//
//   backend {
//     address = "%s"
//     name    = "%s"
//     port    = 443
//   }
//
//   force_destroy = true
// }`, serviceName, domainName, backendAddress, backendName)
// }
//
// // NOTE: We set the port to 443 so we can validate the API is expecting SSL/TLS
// // related attributes to not be accidentally sent with empty strings.
// func testAccServiceVCLBackendUpdate(serviceName, domainName, backendAddress, backendName string) string {
// 	return fmt.Sprintf(`
// resource "fastly_service_vcl" "foo" {
//   name = "%s"
//
//   domain {
//     name    = "%s"
//     comment = "demo"
//   }
//
//   backend {
//     address = "%s"
//     name    = "%s"
//     port    = 443
//   }
//
//   backend {
//     address = "%s"
//     name    = "%s new"
//     port    = 443
//   }
//
//   backend {
//     address           = "%s"
//     name              = "%s new with use ssl"
//     use_ssl           = true
//     ssl_cert_hostname = "httpbin.org"
//   }
//
//   force_destroy = true
// }`, serviceName, domainName, backendAddress, backendName, backendAddress, backendName, backendAddress, backendName)
// }
//
// func testAccCheckFastlyServiceVCLBackendAttributes(service *gofastly.ServiceDetail, want []*gofastly.Backend) resource.TestCheckFunc {
// 	return func(_ *terraform.State) error {
// 		conn := testAccProvider.Meta().(*APIClient).conn
// 		have, err := conn.ListBackends(&gofastly.ListBackendsInput{
// 			ServiceID:      service.ID,
// 			ServiceVersion: service.ActiveVersion.Number,
// 		})
// 		if err != nil {
// 			return fmt.Errorf("error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
// 		}
//
// 		if len(have) != len(want) {
// 			return fmt.Errorf("backend list count mismatch, expected (%d), got (%d)", len(want), len(have))
// 		}
//
// 		var found int
// 		for _, w := range want {
// 			for _, h := range have {
// 				if w.Name == h.Name {
// 					// we don't know these things ahead of time, so populate them now
// 					w.ServiceID = service.ID
// 					w.ServiceVersion = service.ActiveVersion.Number
// 					// We don't track these, so clear them out because we also won't know
// 					// these ahead of time
// 					h.CreatedAt = nil
// 					h.UpdatedAt = nil
// 					if !reflect.DeepEqual(w, h) {
// 						return fmt.Errorf("bad match Backend match, expected (%#v), got (%#v)", w, h)
// 					}
// 					found++
// 				}
// 			}
// 		}
//
// 		if found != len(want) {
// 			return fmt.Errorf("error matching Backends (%d/%d)", found, len(want))
// 		}
//
// 		return nil
// 	}
// }
