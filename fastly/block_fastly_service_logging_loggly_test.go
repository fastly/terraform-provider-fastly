package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenLoggly(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Loggly
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Loggly{
				{
					ServiceVersion: gofastly.ToPointer(1),
					Name:           gofastly.ToPointer("loggly-endpoint"),
					Token:          gofastly.ToPointer("token"),
					FormatVersion:  gofastly.ToPointer(2),
				},
			},
			local: []map[string]any{
				{
					"name":           "loggly-endpoint",
					"token":          "token",
					"format_version": 2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenLoggly(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCL_logging_loggly_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Loggly{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("loggly-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("s3cr3t"),
	}

	log1AfterUpdate := gofastly.Loggly{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b %T"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("loggly-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("secret"),
	}

	log2 := gofastly.Loggly{
		Format:            gofastly.ToPointer("%h %l %u %t \"%r\" %>s %b"),
		FormatVersion:     gofastly.ToPointer(2),
		Name:              gofastly.ToPointer("another-loggly-endpoint"),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Token:             gofastly.ToPointer("another-token"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogglyConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogglyAttributes(&service, []*gofastly.Loggly{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_loggly.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLLogglyConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogglyAttributes(&service, []*gofastly.Loggly{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "logging_loggly.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logging_loggly_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Loggly{
		Name:           gofastly.ToPointer("loggly-endpoint"),
		ServiceVersion: gofastly.ToPointer(1),
		Token:          gofastly.ToPointer("s3cr3t"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogglyComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLLogglyAttributes(&service, []*gofastly.Loggly{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_compute.foo", "logging_loggly.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLLogglyAttributes(service *gofastly.ServiceDetail, loggly []*gofastly.Loggly, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		logglyList, err := conn.ListLoggly(&gofastly.ListLogglyInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Loggly Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(logglyList) != len(loggly) {
			return fmt.Errorf("loggly List count mismatch, expected (%d), got (%d)", len(loggly), len(logglyList))
		}

		log.Printf("[DEBUG] logglyList = %#v\n", logglyList)

		var found int
		for _, e := range loggly {
			for _, el := range logglyList {
				if gofastly.ToValue(e.Name) == gofastly.ToValue(el.Name) {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					el.CreatedAt = nil
					el.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						el.FormatVersion = e.FormatVersion
						el.Format = e.Format
						el.ResponseCondition = e.ResponseCondition
						el.Placement = e.Placement
					}

					if diff := cmp.Diff(e, el); diff != "" {
						return fmt.Errorf("bad match Loggly logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(loggly) {
			return fmt.Errorf("error matching Loggly Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLLogglyComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-loggly-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_loggly {
    name   = "loggly-endpoint"
    token  = "s3cr3t"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLLogglyConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-loggly-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_loggly {
    name   = "loggly-endpoint"
    token  = "s3cr3t"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLLogglyConfigUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-loggly-logging"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_loggly {
    name   = "loggly-endpoint"
    token  = "secret"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b %%T"
  }

  logging_loggly {
    name   = "another-loggly-endpoint"
    token  = "another-token"
    format = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
  }

  force_destroy = true
}
`, name, domain)
}
