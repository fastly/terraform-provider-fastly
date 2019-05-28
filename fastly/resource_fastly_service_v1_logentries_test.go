package fastly

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccFastlyServiceV1_logentries_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Logentries{
		Version:           1,
		Name:              "somelogentriesname",
		Port:              uint(20000),
		UseTLS:            true,
		Token:             "token",
		Format:            "%h %l %u %t %r %>s",
		FormatVersion:     1,
		ResponseCondition: "response_condition_test",
	}

	log2 := gofastly.Logentries{
		Version:           1,
		Name:              "somelogentriesanothername",
		Port:              uint(10000),
		UseTLS:            false,
		Token:             "newtoken",
		Format:            "%h %u %t %r %>s",
		FormatVersion:     1,
		ResponseCondition: "response_condition_test",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1LogentriesConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogentriesAttributes(&service, []*gofastly.Logentries{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logentries.#", "1"),
				),
			},

			{
				Config: testAccServiceV1LogentriesConfig_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogentriesAttributes(&service, []*gofastly.Logentries{&log1, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logentries.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1LogentriesAttributes(service *gofastly.ServiceDetail, logentriess []*gofastly.Logentries) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		logentriesList, err := conn.ListLogentries(&gofastly.ListLogentriesInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Logentries Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(logentriesList) != len(logentriess) {
			return fmt.Errorf("Logentries List count mismatch, expected (%d), got (%d)", len(logentriess), len(logentriesList))
		}

		log.Printf("[DEBUG] logentriesList = %+v\n", logentriesList)

		var found int
		for _, s := range logentriess {
			for _, ls := range logentriesList {
				if s.Name == ls.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					ls.CreatedAt = nil
					ls.UpdatedAt = nil
					if !reflect.DeepEqual(s, ls) {
						return fmt.Errorf("Bad match Logentries logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", s, ls)
					}
					found++
				}
			}
		}

		if found != len(logentriess) {
			return fmt.Errorf("Error matching Logentries Logging rules")
		}

		return nil
	}
}

func TestAccFastlyServiceV1_logentries_formatVersion(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Logentries{
		Version:           1,
		Name:              "somelogentriesname",
		Port:              uint(20000),
		UseTLS:            true,
		Token:             "token",
		Format:            "%h %l %u %t %r %>s",
		FormatVersion:     2,
		ResponseCondition: "response_condition_test",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1LogentriesConfig_formatVersion(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogentriesAttributes(&service, []*gofastly.Logentries{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logentries.#", "1"),
				),
			},
		},
	})
}

func testAccServiceV1LogentriesConfig(name, domain string) string {
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
  logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceV1LogentriesConfig_update(name, domain string) string {
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
  logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
  }
  logentries {
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

func testAccServiceV1LogentriesConfig_formatVersion(name, domain string) string {
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
  logentries {
    name               = "somelogentriesname"
    token              = "token"
    response_condition = "response_condition_test"
    format_version     = 2
  }
  force_destroy = true
}`, name, domain)
}
