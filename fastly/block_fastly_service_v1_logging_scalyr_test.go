package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenScalyr(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Scalyr
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Scalyr{
				{
					ServiceVersion:    1,
					Name:              "scalyr-endpoint",
					Region:            "US",
					Token:             "tkn",
					ResponseCondition: "response_condition",
					Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
					FormatVersion:     2,
					Placement:         "none",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "scalyr-endpoint",
					"region":             "US",
					"token":              "tkn",
					"response_condition": "response_condition",
					"format":             `%a %l %u %t %m %U%q %H %>s %b %T`,
					"placement":          "none",
					"format_version":     uint(2),
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

func TestAccFastlyServiceV1_scalyrlogging_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Scalyr{
		ServiceVersion:    1,
		Name:              "scalyrlogger",
		Token:             "tkn",
		Placement:         "none",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		ResponseCondition: "response_condition_test",

		Region:        "US",
		FormatVersion: 2,
	}

	log1_after_update := gofastly.Scalyr{
		ServiceVersion:    1,
		Name:              "scalyrlogger",
		Region:            "EU",
		Token:             "newtkn",
		Placement:         "waf_debug",
		Format:            `%a %l %u %t %m %U%q %H %>s %b %T`,
		ResponseCondition: "response_condition_test",

		FormatVersion: 2,
	}

	log2 := gofastly.Scalyr{
		ServiceVersion: 1,
		Name:           "another-scalyrlogger",
		Token:          "tknb",
		Placement:      "none",
		Format:         `%a %l %u %t %m %U%q %H %>s %b %T`,

		Region:        "US",
		FormatVersion: 2,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{

			{
				Config: testAccServiceV1ScalyrConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ScalyrAttributes(&service, []*gofastly.Scalyr{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_scalyr.#", "1"),
				),
			},

			{
				Config: testAccServiceV1ScalyrConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1ScalyrAttributes(&service, []*gofastly.Scalyr{&log1_after_update, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_scalyr.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_scalyrlogging_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Scalyr{
		ServiceVersion: 1,
		Name:           "scalyrlogger",
		Token:          "tkn",
		Region:         "US",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1ScalyrComputeConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceV1ScalyrAttributes(&service, []*gofastly.Scalyr{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_scalyr.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1ScalyrAttributes(service *gofastly.ServiceDetail, scalyr []*gofastly.Scalyr, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		scalyrList, err := conn.ListScalyrs(&gofastly.ListScalyrsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Scalyr Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(scalyrList) != len(scalyr) {
			return fmt.Errorf("Scalyr List count mismatch, expected (%d), got (%d)", len(scalyr), len(scalyrList))
		}

		log.Printf("[DEBUG] scalyrList = %#v\n", scalyrList)

		var found int
		for _, s := range scalyr {
			for _, sl := range scalyrList {
				if s.Name == sl.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
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
						return fmt.Errorf("Bad match Scalyr logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(scalyr) {
			return fmt.Errorf("Error matching Scalyr Logging rules")
		}

		return nil
	}
}

func testAccServiceV1ScalyrComputeConfig(name string, domain string) string {
	return fmt.Sprintf(`
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
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}

	force_destroy = true
}
`, name, domain)
}

func testAccServiceV1ScalyrConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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

func testAccServiceV1ScalyrConfig_update(name, domain string) string {
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
