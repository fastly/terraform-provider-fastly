package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenPapertrail(t *testing.T) {

	cases := []struct {
		remote []*gofastly.Papertrail
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Papertrail{
				{
					ServiceVersion:    1,
					Name:              "papertrailtesting",
					Address:           "test1.papertrailapp.com",
					Port:              3600,
					Format:            "%h %l %u %t %r %>s",
					FormatVersion:     2,
					ResponseCondition: "test_response_condition",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "papertrailtesting",
					"address":            "test1.papertrailapp.com",
					"port":               uint(3600),
					"format":             "%h %l %u %t %r %>s",
					"format_version":     uint(2),
					"response_condition": "test_response_condition",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenPapertrails(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}

}

func TestAccFastlyServiceVCL_papertrail_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Papertrail{
		ServiceVersion:    1,
		Name:              "papertrailtesting",
		Address:           "test1.papertrailapp.com",
		Port:              uint(3600),
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     uint(2),
		ResponseCondition: "test_response_condition",
	}

	log2 := gofastly.Papertrail{
		ServiceVersion: 1,
		Name:           "papertrailtesting2",
		Address:        "test2.papertrailapp.com",
		Port:           uint(8080),
		Format:         `%h %l %u %t "%r" %>s %b`,
		FormatVersion:  uint(2),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccServiceVCLPapertrailConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLPapertrailAttributes(&service, []*gofastly.Papertrail{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_papertrail.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLPapertrailConfig_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLPapertrailAttributes(&service, []*gofastly.Papertrail{&log1, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_papertrail.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_papertrail_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Papertrail{
		ServiceVersion: 1,
		Name:           "papertrailtesting",
		Address:        "test1.papertrailapp.com",
		Port:           uint(3600),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLPapertrailComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLPapertrailAttributes(&service, []*gofastly.Papertrail{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_papertrail.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLPapertrailAttributes(service *gofastly.ServiceDetail, papertrails []*gofastly.Papertrail, serviceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		papertrailList, err := conn.ListPapertrails(&gofastly.ListPapertrailsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Papertrail for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(papertrailList) != len(papertrails) {
			return fmt.Errorf("Papertrail List count mismatch, expected (%d), got (%d)", len(papertrails), len(papertrailList))
		}

		var found int
		for _, p := range papertrails {
			for _, lp := range papertrailList {
				if p.Name == lp.Name {
					// we don't know these things ahead of time, so populate them now
					p.ServiceID = service.ID
					p.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lp.CreatedAt = nil
					lp.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						lp.FormatVersion = p.FormatVersion
						lp.Format = p.Format
						lp.ResponseCondition = p.ResponseCondition
						lp.Placement = p.Placement
					}

					if !reflect.DeepEqual(p, lp) {
						return fmt.Errorf("Bad match Papertrail match, expected (%#v), got (%#v)", p, lp)
					}
					found++
				}
			}
		}

		if found != len(papertrails) {
			return fmt.Errorf("Error matching Papertrail rules")
		}

		return nil
	}
}

func testAccServiceVCLPapertrailComputeConfig(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  logging_papertrail {
    name               = "papertrailtesting"
    address            = "test1.papertrailapp.com"
    port               = 3600
  }

  package {
      	filename = "test_fixtures/package/valid.tar.gz"
	  	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
   	}

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLPapertrailConfig(name, domain string) string {
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
    name      = "test_response_condition"
    type      = "RESPONSE"
    priority  = 5
    statement = "resp.status >= 400 && resp.status < 600"
  }

  logging_papertrail {
    name               = "papertrailtesting"
    address            = "test1.papertrailapp.com"
    port               = 3600
		response_condition = "test_response_condition"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLPapertrailConfig_update(name, domain string) string {
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
    name      = "test_response_condition"
    type      = "RESPONSE"
    priority  = 5
    statement = "resp.status >= 400 && resp.status < 600"
  }

	logging_papertrail {
    name               = "papertrailtesting"
    address            = "test1.papertrailapp.com"
    port               = 3600
		response_condition = "test_response_condition"
  }

	logging_papertrail {
    name               = "papertrailtesting2"
    address            = "test2.papertrailapp.com"
    port               = 8080
  }

  force_destroy = true
}`, name, domain)
}
