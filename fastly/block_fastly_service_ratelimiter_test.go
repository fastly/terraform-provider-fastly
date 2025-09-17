package fastly

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
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
					Action: gofastly.ToPointer(gofastly.ERLActionResponse),
					ClientKey: []*string{
						gofastly.ToPointer("req.http.Fastly-Client-IP"),
						gofastly.ToPointer("req.http.User-Agent"),
					},
					FeatureRevision: gofastly.ToPointer(1),
					HTTPMethods: []*string{
						gofastly.ToPointer("POST"),
						gofastly.ToPointer("PUT"),
						gofastly.ToPointer("PATCH"),
						gofastly.ToPointer("DELETE"),
					},
					RateLimiterID:      gofastly.ToPointer("123abc"),
					LoggerType:         gofastly.ToPointer(gofastly.ERLLogBigQuery),
					Name:               gofastly.ToPointer("example"),
					PenaltyBoxDuration: gofastly.ToPointer(123),
					Response: &gofastly.ERLResponse{
						ERLContent:     gofastly.ToPointer("example"),
						ERLContentType: gofastly.ToPointer("plain/text"),
						ERLStatus:      gofastly.ToPointer(429),
					},
					ResponseObjectName: gofastly.ToPointer("example"),
					RpsLimit:           gofastly.ToPointer(123),
					WindowSize:         gofastly.ToPointer(gofastly.ERLSize1),
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

func TestAccFastlyServiceVCLRateLimiter_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	rateLimiterName := fmt.Sprintf("backend-tf-%s", acctest.RandString(10))

	// The following Rate Limiters are what we expect to exist after all our
	// Terraform configuration settings have been applied. We expect them to
	// correlate to the specific Rate Limiter definitions in the Terraform config.

	erl1 := gofastly.ERL{
		Action: gofastly.ToPointer(gofastly.ERLActionResponse),
		ClientKey: []*string{
			gofastly.ToPointer("req.http.Fastly-Client-IP"),
			gofastly.ToPointer("req.http.User-Agent"),
		},
		FeatureRevision: gofastly.ToPointer(1),
		HTTPMethods: []*string{
			gofastly.ToPointer("POST"),
			gofastly.ToPointer("PUT"),
			gofastly.ToPointer("PATCH"),
			gofastly.ToPointer("DELETE"),
		},
		Name:               gofastly.ToPointer(rateLimiterName),
		PenaltyBoxDuration: gofastly.ToPointer(30),
		Response: &gofastly.ERLResponse{
			ERLContent:     gofastly.ToPointer("example"),
			ERLContentType: gofastly.ToPointer("plain/text"),
			ERLStatus:      gofastly.ToPointer(429),
		},
		RpsLimit:   gofastly.ToPointer(100),
		WindowSize: gofastly.ToPointer(gofastly.ERLSize60),
	}

	erl1Updated := gofastly.ERL{
		Action: gofastly.ToPointer(gofastly.ERLActionResponse),
		ClientKey: []*string{
			gofastly.ToPointer("req.http.Fastly-Client-IP"),
			gofastly.ToPointer("req.http.User-Agent"),
		},
		FeatureRevision: gofastly.ToPointer(1),
		HTTPMethods: []*string{
			gofastly.ToPointer("POST"),
			gofastly.ToPointer("PUT"),
			gofastly.ToPointer("PATCH"),
			gofastly.ToPointer("DELETE"),
		},
		Name:               gofastly.ToPointer(rateLimiterName),
		PenaltyBoxDuration: gofastly.ToPointer(31),
		Response: &gofastly.ERLResponse{
			ERLContent:     gofastly.ToPointer("example"),
			ERLContentType: gofastly.ToPointer("plain/text"),
			ERLStatus:      gofastly.ToPointer(429),
		},
		RpsLimit:          gofastly.ToPointer(100),
		WindowSize:        gofastly.ToPointer(gofastly.ERLSize60),
		URIDictionaryName: gofastly.ToPointer("rate_limit_endpoints"),
	}

	erl2 := gofastly.ERL{
		Action: gofastly.ToPointer(gofastly.ERLActionResponse),
		ClientKey: []*string{
			gofastly.ToPointer("req.http.Fastly-Client-IP"),
			gofastly.ToPointer("req.http.User-Agent"),
		},
		FeatureRevision: gofastly.ToPointer(1),
		HTTPMethods: []*string{
			gofastly.ToPointer("POST"),
			gofastly.ToPointer("PUT"),
			gofastly.ToPointer("PATCH"),
			gofastly.ToPointer("DELETE"),
		},
		Name:               gofastly.ToPointer(rateLimiterName + "-2"),
		PenaltyBoxDuration: gofastly.ToPointer(30),
		Response: &gofastly.ERLResponse{
			ERLContent:     gofastly.ToPointer("example"),
			ERLContentType: gofastly.ToPointer("plain/text"),
			ERLStatus:      gofastly.ToPointer(429),
		},
		RpsLimit:   gofastly.ToPointer(100),
		WindowSize: gofastly.ToPointer(gofastly.ERLSize60),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLRateLimiter(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.example", &service),
					resource.TestCheckResourceAttr("fastly_service_vcl.example", "name", serviceName),
					resource.TestCheckResourceAttr("fastly_service_vcl.example", "rate_limiter.#", "1"),
					testAccCheckFastlyServiceVCLRateLimiterAttributes(&service, []*gofastly.ERL{&erl1}),
				),
			},

			{
				Config: testAccServiceVCLRateLimiterUpdate(serviceName, domainName, rateLimiterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.example", &service),
					testAccCheckRateLimiterURIDictName("fastly_service_vcl.example"),
					resource.TestCheckResourceAttr("fastly_service_vcl.example", "rate_limiter.#", "2"),
					testAccCheckFastlyServiceVCLRateLimiterAttributes(&service, []*gofastly.ERL{&erl1Updated, &erl2}),
				),
			},

			{
				Config:      testAccServiceVCLMultipleRateLimiters(serviceName, domainName, rateLimiterName),
				ExpectError: regexp.MustCompile("multiple rate_limiters with the same name"),
			},
		},
	})
}

func testAccServiceVCLRateLimiter(serviceName, domainName, rateLimiterName string) string {
	return fmt.Sprintf(`
variable "mydict" {
  type = string
  default = "rate_limit_endpoints"
}

resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  dictionary {
    name = var.mydict
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s"
    penalty_box_duration = 30

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 100
    window_size = 60
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items" "rate_limit_endpoints" {
  service_id = fastly_service_vcl.example.id
  dictionary_id = { for d in fastly_service_vcl.example.dictionary : d.name => d.dictionary_id }[var.mydict]
  items = {
    key1: "value1"
    key2: "value2"
  }
}`, serviceName, domainName, rateLimiterName)
}

// The following config...
//
// Updates:
// - penalty_box_duration (30 -> 31)
//
// Adds:
// - A second rate_limiter
// - uri_dictionary_name (to the first rate_limter).
func testAccServiceVCLRateLimiterUpdate(serviceName, domainName, rateLimiterName string) string {
	return fmt.Sprintf(`
variable "mydict" {
  type = string
  default = "rate_limit_endpoints"
}

resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  dictionary {
    name = var.mydict
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s"
    penalty_box_duration = 31
    uri_dictionary_name  = var.mydict

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 100
    window_size = 60
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s-2"
    penalty_box_duration = 30

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 100
    window_size = 60
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items" "rate_limit_endpoints" {
  service_id = fastly_service_vcl.example.id
  dictionary_id = { for d in fastly_service_vcl.example.dictionary : d.name => d.dictionary_id }[var.mydict]
  items = {
    key1: "value1"
    key2: "value2"
  }
}`, serviceName, domainName, rateLimiterName, rateLimiterName)
}

// IMPORTANT: The following config defines two rate limiters with the same 'name'.
// Although allowed by the Fastly API, this isn't ideal.
// That's because we need the names to be unique for the purpose of updating
// rate limiters (as the API also causes each Rate Limiter's ID to change when a
// service is cloned).
// The Fastly Terraform provider should return an error when generating the diff.
func testAccServiceVCLMultipleRateLimiters(serviceName, domainName, rateLimiterName string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "example" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s"
    penalty_box_duration = 30

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 100
    window_size = 60
  }

  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s-2"
    penalty_box_duration = 30

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 100
    window_size = 60
  }

  # IMPORTANT: The following Rate Limiter has the same 'name' as above.
  # But to ensure the error is reported we have to have some other difference.
  # In this case we change the rps_limit.
  # If we just took an exact copy of the above rate limiter, then we'd have no
  # error reported from Terraform because the TypeSet behaviour would kick in
  # and prevent the plan/diff from even showing any change because a set data
  # structure doesn't allow any duplicates (hence we need a small difference
  # just for the sake of validating the 'name' fields don't match).
  rate_limiter {
    action               = "response"
    client_key           = "req.http.Fastly-Client-IP,req.http.User-Agent"
    http_methods         = "POST,PUT,PATCH,DELETE"
    name                 = "%s-2"
    penalty_box_duration = 30

    response {
      content      = "example"
      content_type = "plain/text"
      status       = 429
    }

    rps_limit   = 1
    window_size = 60
  }

  force_destroy = true
}`, serviceName, domainName, rateLimiterName, rateLimiterName, rateLimiterName)
}

func testAccCheckFastlyServiceVCLRateLimiterAttributes(service *gofastly.ServiceDetail, want []*gofastly.ERL) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		have, err := conn.ListERLs(context.TODO(), &gofastly.ListERLsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Rate Limiters for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(have) != len(want) {
			return fmt.Errorf("backend list count mismatch, expected (%d), got (%d)", len(want), len(have))
		}

		var found int
		for _, w := range want {
			for _, h := range have {
				if gofastly.ToValue(w.Name) == gofastly.ToValue(h.Name) {
					// we don't know these things ahead of time, so populate them now
					w.RateLimiterID = h.RateLimiterID
					w.ServiceID = service.ServiceID
					w.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					h.CreatedAt = nil
					h.UpdatedAt = nil
					if !reflect.DeepEqual(w, h) {
						return fmt.Errorf("bad match Rate Limiters match, expected (%#v), got (%#v)", w, h)
					}
					found++
				}
			}
		}

		if found != len(want) {
			return fmt.Errorf("error matching Rate Limiters (%d/%d)", found, len(want))
		}

		return nil
	}
}

// testAccCheckRateLimiterURIDictName validates uri_dictionary_name is set with
// a value after it is defined in the Terraform config.
func testAccCheckRateLimiterURIDictName(vclServiceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[vclServiceName]
		if !ok {
			return fmt.Errorf("not found: %s", vclServiceName)
		}
		attrs := rs.Primary.Attributes
		a := "rate_limiter.0.uri_dictionary_name"
		v, ok := attrs[a]
		if !ok {
			return fmt.Errorf("failed to lookup %s", a)
		}
		if v == "" {
			return fmt.Errorf("%s is empty", a)
		}
		return nil
	}
}
