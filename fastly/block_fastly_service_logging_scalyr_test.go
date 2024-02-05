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

func TestResourceFastlyFlattenScalyr(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Scalyr
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Scalyr{
				{
					ServiceVersion:    gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("scalyr-endpoint"),
					Region:            gofastly.ToPointer("US"),
					Token:             gofastly.ToPointer("tkn"),
					ResponseCondition: gofastly.ToPointer("response_condition"),
					Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
					FormatVersion:     gofastly.ToPointer(2),
					Placement:         gofastly.ToPointer("none"),
				},
			},
			local: []map[string]any{
				{
					"name":               "scalyr-endpoint",
					"region":             "US",
					"token":              "tkn",
					"response_condition": "response_condition",
					"format":             `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":          "none",
					"format_version":     2,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenScalyr(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}

func TestAccFastlyServiceVCL_scalyrlogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Scalyr{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("scalyrlogger"),
		Token:             gofastly.ToPointer("tkn"),
		Placement:         gofastly.ToPointer("none"),
		Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		Region:            gofastly.ToPointer("US"),
		FormatVersion:     gofastly.ToPointer(2),
	}

	log1AfterUpdate := gofastly.Scalyr{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("scalyrlogger"),
		Region:            gofastly.ToPointer("EU"),
		Token:             gofastly.ToPointer("newtkn"),
		Placement:         gofastly.ToPointer("waf_debug"),
		Format:            gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		FormatVersion:     gofastly.ToPointer(2),
	}

	log2 := gofastly.Scalyr{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("another-scalyrlogger"),
		Token:          gofastly.ToPointer("tknb"),
		Placement:      gofastly.ToPointer("none"),
		Format:         gofastly.ToPointer(`%a %l %u %t %m %U%q %H %>s %b %T`),
		Region:         gofastly.ToPointer("US"),
		FormatVersion:  gofastly.ToPointer(2),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLScalyrConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLScalyrAttributes(&service, []*gofastly.Scalyr{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_scalyr.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLScalyrConfigUpdate(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLScalyrAttributes(&service, []*gofastly.Scalyr{&log1AfterUpdate, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_scalyr.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_scalyrlogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Scalyr{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("scalyrlogger"),
		Token:          gofastly.ToPointer("tkn"),
		Region:         gofastly.ToPointer("US"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLScalyrComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLScalyrAttributes(&service, []*gofastly.Scalyr{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_scalyr.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLScalyrAttributes(service *gofastly.ServiceDetail, scalyr []*gofastly.Scalyr, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		scalyrList, err := conn.ListScalyrs(&gofastly.ListScalyrsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Scalyr Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(scalyrList) != len(scalyr) {
			return fmt.Errorf("scalyr List count mismatch, expected (%d), got (%d)", len(scalyr), len(scalyrList))
		}

		log.Printf("[DEBUG] scalyrList = %#v\n", scalyrList)

		var found int
		for _, s := range scalyr {
			for _, sl := range scalyrList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(sl.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					sl.CreatedAt = nil
					sl.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						sl.FormatVersion = s.FormatVersion
						sl.Format = s.Format
						sl.ResponseCondition = s.ResponseCondition
						sl.Placement = s.Placement
					}

					if diff := cmp.Diff(s, sl); diff != "" {
						return fmt.Errorf("bad match Scalyr logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(scalyr) {
			return fmt.Errorf("error matching Scalyr Logging rules")
		}

		return nil
	}
}

func testAccServiceVCLScalyrComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

resource "fastly_service_compute" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-scalyr-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	logging_scalyr {
		name               = "scalyrlogger"
		region             = "US"
		token              = "tkn"
	}

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLScalyrConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
	name = "%s"

	domain {
		name    = "%s"
		comment = "tf-scalyr-logging"
	}

	backend {
		address = "aws.amazon.com"
		name    = "amazon docs"
	}

	condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

	logging_scalyr {
		name               = "scalyrlogger"
		region             = "US"
		token              = "tkn"
		response_condition = "response_condition_test"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version 		 = 2
		placement 				 = "none"
	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceVCLScalyrConfigUpdate(name, domain string) string {
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

	condition {
    name      = "response_condition_test"
    type      = "RESPONSE"
    priority  = 8
    statement = "resp.status == 418"
  }

	logging_scalyr {
		name               = "scalyrlogger"
		region             = "EU"
		token              = "newtkn"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		format_version 		 = 2
		response_condition = "response_condition_test"
		placement 				 = "waf_debug"
	}

	logging_scalyr {
		name               = "another-scalyrlogger"
		format             = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
		region             = "US"
		token              = "tknb"
		format_version 		 = 2
		placement 				 = "none"
	}

	force_destroy = true
}`, name, domain)
}
