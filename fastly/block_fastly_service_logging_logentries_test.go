package fastly

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
)

func TestResourceFastlyFlattenLogentries(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Logentries
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Logentries{
				{
					Format:            gofastly.ToPointer("%h %l %u %t %r %>s"),
					FormatVersion:     gofastly.ToPointer(1),
					Name:              gofastly.ToPointer("somelogentriesname"),
					Placement:         gofastly.ToPointer("placement"),
					Port:              gofastly.ToPointer(8080),
					ResponseCondition: gofastly.ToPointer("response_condition_test"),
					ServiceVersion:    gofastly.ToPointer(1), // expect this not to be persisted to tf state as it's tracked by the parent 'service' resource
					Token:             gofastly.ToPointer("mytoken"),
					ProcessingRegion:  gofastly.ToPointer("eu"),
				},
			},
			local: []map[string]any{
				{
					"format":             "%h %l %u %t %r %>s",
					"format_version":     1,
					"name":               "somelogentriesname",
					"placement":          "placement",
					"port":               8080,
					"response_condition": "response_condition_test",
					"token":              "mytoken",
					"processing_region":  "eu",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenLogentries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_logentries_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Logentries{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somelogentriesname"),
		Port:              gofastly.ToPointer(20000),
		UseTLS:            gofastly.ToPointer(true),
		Token:             gofastly.ToPointer("token"),
		Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	log2 := gofastly.Logentries{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somelogentriesanothername"),
		Port:              gofastly.ToPointer(10000),
		UseTLS:            gofastly.ToPointer(false),
		Token:             gofastly.ToPointer("newtoken"),
		Format:            gofastly.ToPointer("%h %u %t %r %>s"),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ProcessingRegion:  gofastly.ToPointer("none"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogentriesConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogentriesAttributes(&service, []*gofastly.Logentries{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_logentries.#", "1"),
				),
			},
			{
				Config: testAccServiceVCLLogentriesConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogentriesAttributes(&service, []*gofastly.Logentries{&log1, &log2}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_logentries.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceVCL_logentries_basic_compute(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Logentries{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somelogentriesname"),
		Port:              gofastly.ToPointer(20000),
		UseTLS:            gofastly.ToPointer(true),
		Token:             gofastly.ToPointer("token"),
		Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
		ProcessingRegion:  gofastly.ToPointer("us"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogentriesComputeConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_compute.foo", &service),
					testAccCheckFastlyServiceVCLLogentriesAttributes(&service, []*gofastly.Logentries{&log1}, ServiceTypeCompute),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_compute.foo", "logging_logentries.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLLogentriesAttributes(service *gofastly.ServiceDetail, logentriess []*gofastly.Logentries, serviceType string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		logentriesList, err := conn.ListLogentries(context.TODO(), &gofastly.ListLogentriesInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Logentries Logging for (%s), version (%d): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(logentriesList) != len(logentriess) {
			return fmt.Errorf("logentries List count mismatch, expected (%d), got (%d)", len(logentriess), len(logentriesList))
		}

		log.Printf("[DEBUG] logentriesList = %+v\n", logentriesList)

		var found int
		for _, s := range logentriess {
			for _, ls := range logentriesList {
				if gofastly.ToValue(s.Name) == gofastly.ToValue(ls.Name) {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ServiceID
					s.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					ls.CreatedAt = nil
					ls.UpdatedAt = nil

					// Ignore VCL attributes for Compute and set to whatever is returned from the API.
					if serviceType == ServiceTypeCompute {
						ls.FormatVersion = s.FormatVersion
						ls.Format = s.Format
						ls.ResponseCondition = s.ResponseCondition
						ls.Placement = s.Placement
					}

					if !reflect.DeepEqual(s, ls) {
						return fmt.Errorf("bad match Logentries logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", s, ls)
					}
					found++
				}
			}
		}

		if found != len(logentriess) {
			return fmt.Errorf("error matching Logentries Logging rules")
		}

		return nil
	}
}

func TestAccFastlyServiceVCL_logentries_formatVersion(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Logentries{
		ServiceVersion:    gofastly.ToPointer(1),
		Name:              gofastly.ToPointer("somelogentriesname"),
		Port:              gofastly.ToPointer(20000),
		UseTLS:            gofastly.ToPointer(true),
		Token:             gofastly.ToPointer("token"),
		Format:            gofastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     gofastly.ToPointer(2),
		ResponseCondition: gofastly.ToPointer("response_condition_test"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLLogentriesConfigFormatVersion(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLLogentriesAttributes(&service, []*gofastly.Logentries{&log1}, ServiceTypeVCL),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "logging_logentries.#", "1"),
				),
			},
		},
	})
}

func testAccServiceVCLLogentriesComputeConfig(name, domain string) string {
	return fmt.Sprintf(`
data "fastly_package_hash" "example" {
  filename = "./test_fixtures/package/valid.tar.gz"
}

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

  logging_logentries {
    name               = "somelogentriesname"
    token              = "token"
    processing_region = "us"
  }

  package {
    filename = "test_fixtures/package/valid.tar.gz"
    source_code_hash = data.fastly_package_hash.example.hash
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLLogentriesConfig(name, domain string) string {
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
  logging_logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
    processing_region = "us"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLLogentriesConfigUpdate(name, domain string) string {
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
  logging_logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
  }
  logging_logentries {
    name               = "somelogentriesanothername"
    port               = "10000"
    use_tls            = "false"
    token              = "newtoken"
    format             = "%%h %%u %%t %%r %%>s"
    response_condition = "response_condition_test"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLLogentriesConfigFormatVersion(name, domain string) string {
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
  logging_logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
    format_version     = 2
  }
  force_destroy = true
}`, name, domain)
}
