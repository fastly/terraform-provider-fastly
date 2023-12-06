package fastly

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenHealthChecks(t *testing.T) {
	cases := []struct {
		remote []*gofastly.HealthCheck
		local  []map[string]any
	}{
		{
			remote: []*gofastly.HealthCheck{
				{
					ServiceVersion:   gofastly.ToPointer(1),
					Name:             gofastly.ToPointer("myhealthcheck"),
					Headers:          []string{"Foo: Bar", "Baz: Qux"},
					Host:             gofastly.ToPointer("example1.com"),
					Path:             gofastly.ToPointer("/test1.txt"),
					CheckInterval:    gofastly.ToPointer(4000),
					ExpectedResponse: gofastly.ToPointer(200),
					HTTPVersion:      gofastly.ToPointer("1.1"),
					Initial:          gofastly.ToPointer(2),
					Method:           gofastly.ToPointer("HEAD"),
					Threshold:        gofastly.ToPointer(3),
					Timeout:          gofastly.ToPointer(5000),
					Window:           gofastly.ToPointer(5),
				},
			},
			local: []map[string]any{
				{
					"name":              "myhealthcheck",
					"headers":           []string{"Foo: Bar", "Baz: Qux"},
					"host":              "example1.com",
					"path":              "/test1.txt",
					"check_interval":    4000,
					"expected_response": 200,
					"http_version":      "1.1",
					"initial":           2,
					"method":            "HEAD",
					"threshold":         3,
					"timeout":           5000,
					"window":            5,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHealthchecks(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_healthcheck_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.HealthCheck{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("example-healthcheck1"),
		Headers:          []string{"Foo: Bar", "Baz: Qux"},
		Host:             gofastly.ToPointer("example1.com"),
		Path:             gofastly.ToPointer("/test1.txt"),
		CheckInterval:    gofastly.ToPointer(4000),
		ExpectedResponse: gofastly.ToPointer(200),
		HTTPVersion:      gofastly.ToPointer("1.1"),
		Initial:          gofastly.ToPointer(2),
		Method:           gofastly.ToPointer("HEAD"),
		Threshold:        gofastly.ToPointer(3),
		Timeout:          gofastly.ToPointer(5000),
		Window:           gofastly.ToPointer(5),
	}

	log2 := gofastly.HealthCheck{
		CheckInterval:    gofastly.ToPointer(4500),
		ExpectedResponse: gofastly.ToPointer(404),
		HTTPVersion:      gofastly.ToPointer("1.0"),
		Headers:          []string{"Beep: Boop"},
		Host:             gofastly.ToPointer("example2.com"),
		Initial:          gofastly.ToPointer(1),
		Method:           gofastly.ToPointer("POST"),
		Name:             gofastly.ToPointer("example-healthcheck2"),
		Path:             gofastly.ToPointer("/test2.txt"),
		ServiceVersion:   gofastly.ToPointer(1),
		Threshold:        gofastly.ToPointer(4),
		Timeout:          gofastly.ToPointer(4000),
		Window:           gofastly.ToPointer(10),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHealthCheckConfig(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHealthCheckAttributes(&service, []*gofastly.HealthCheck{&log1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "healthcheck.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLHealthCheckConfigUpdate(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHealthCheckAttributes(&service, []*gofastly.HealthCheck{&log1, &log2}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "healthcheck.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLHealthCheckAttributes(service *gofastly.ServiceDetail, healthchecks []*gofastly.HealthCheck) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		healthcheckList, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Healthcheck for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(healthcheckList) != len(healthchecks) {
			return fmt.Errorf("healthcheck List count mismatch, expected (%d), got (%d)", len(healthchecks), len(healthcheckList))
		}

		var found int
		for _, h := range healthchecks {
			for _, lh := range healthcheckList {
				if gofastly.ToValue(h.Name) == gofastly.ToValue(lh.Name) {
					// The API returns the headers sorted, so to avoid potential errors in
					// the test setup we will order the headers too before comparing.
					//
					// NOTE: Sorting the headers isn't necessary outside of the tests.
					sort.Strings(h.Headers)

					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ID
					h.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lh.CreatedAt = nil
					lh.UpdatedAt = nil
					if !reflect.DeepEqual(h, lh) {
						return fmt.Errorf("bad match Healthcheck match, expected (%#v), got (%#v)", h, lh)
					}
					found++
				}
			}
		}

		if found != len(healthchecks) {
			return fmt.Errorf("error matching Healthcheck rules")
		}

		return nil
	}
}

func testAccServiceVCLHealthCheckConfig(name, domain string) string {
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

  healthcheck {
		check_interval    = 4000
		expected_response = 200
		headers           = ["Foo: Bar", "Baz: Qux"]
		host              = "example1.com"
		http_version      = "1.1"
		initial           = 2
		method            = "HEAD"
		name              = "example-healthcheck1"
		path              = "/test1.txt"
		threshold         = 3
		timeout           = 5000
		window            = 5
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLHealthCheckConfigUpdate(name, domain string) string {
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

	healthcheck {
		check_interval    = 4000
		expected_response = 200
		headers           = ["Foo: Bar", "Baz: Qux"]
		host              = "example1.com"
		http_version      = "1.1"
		initial           = 2
		method            = "HEAD"
		name              = "example-healthcheck1"
		path              = "/test1.txt"
		threshold         = 3
		timeout           = 5000
		window            = 5
  }

	healthcheck {
		check_interval    = 4500
		expected_response = 404
		headers           = ["Beep: Boop"]
		host              = "example2.com"
		http_version      = "1.0"
		initial           = 1
		method            = "POST"
		name              = "example-healthcheck2"
		path              = "/test2.txt"
		threshold         = 4
		timeout           = 4000
		window            = 10
  }

  force_destroy = true
}`, name, domain)
}
