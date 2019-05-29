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

func TestResourceFastlyFlattenSyslog(t *testing.T) {

	cases := []struct {
		remote []*gofastly.Syslog
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Syslog{
				{
					Version:           1,
					Name:              "somesyslogname",
					Address:           "127.0.0.1",
					IPV4:              "127.0.0.1",
					Port:              8080,
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     1,
					ResponseCondition: "response_condition_test",
					MessageType:       "classic",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "somesyslogname",
					"address":            "127.0.0.1",
					"port":               uint(8080),
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(1),
					"response_condition": "response_condition_test",
					"message_type":       "classic",
					"use_tls":            false,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSyslogs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}

}

func TestAccFastlyServiceV1_syslog_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("%s.notadomain1.com", acctest.RandString(10))

	log1 := gofastly.Syslog{
		Version:           1,
		Name:              "somesyslogname",
		Address:           "127.0.0.1",
		IPV4:              "127.0.0.1",
		Port:              uint(514),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		ResponseCondition: "response_condition_test",
		MessageType:       "classic",
	}

	log1_after_update := gofastly.Syslog{
		Version:           1,
		Name:              "somesyslogname",
		Address:           "127.0.0.1",
		IPV4:              "127.0.0.1",
		Port:              uint(514),
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		ResponseCondition: "response_condition_test",
		MessageType:       "blank",
	}

	log2 := gofastly.Syslog{
		Version:       1,
		Name:          "somesyslogname2",
		Address:       "127.0.0.2",
		IPV4:          "127.0.0.2",
		Port:          uint(10514),
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 1,
		MessageType:   "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SyslogConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SyslogAttributes(&service, []*gofastly.Syslog{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "syslog.#", "1"),
				),
			},

			{
				Config: testAccServiceV1SyslogConfig_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SyslogAttributes(&service, []*gofastly.Syslog{&log1_after_update, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "syslog.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_syslog_formatVersion(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("%s.notadomain1.com", acctest.RandString(10))

	log1 := gofastly.Syslog{
		Version:       1,
		Name:          "somesyslogname",
		Address:       "127.0.0.1",
		IPV4:          "127.0.0.1",
		Port:          uint(514),
		Format:        "%a %l %u %t %m %U%q %H %>s %b %T",
		FormatVersion: 2,
		MessageType:   "classic",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SyslogConfig_formatVersion(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SyslogAttributes(&service, []*gofastly.Syslog{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "syslog.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1SyslogAttributes(service *gofastly.ServiceDetail, syslogs []*gofastly.Syslog) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		syslogList, err := conn.ListSyslogs(&gofastly.ListSyslogsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Syslog Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(syslogList) != len(syslogs) {
			return fmt.Errorf("Syslog List count mismatch, expected (%d), got (%d)", len(syslogs), len(syslogList))
		}

		log.Printf("[DEBUG] syslogList = %+v\n", syslogList)

		var found int
		for _, s := range syslogs {
			for _, ls := range syslogList {
				if s.Name == ls.Name {
					// we don't know these things ahead of time, so populate them now
					s.ServiceID = service.ID
					s.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					ls.CreatedAt = nil
					ls.UpdatedAt = nil
					if !reflect.DeepEqual(s, ls) {
						return fmt.Errorf("Bad match Syslog logging match,\nexpected:\n(%#v),\ngot:\n(%#v)", s, ls)
					}
					found++
				}
			}
		}

		if found != len(syslogs) {
			return fmt.Errorf("Error matching Syslog Logging rules")
		}

		return nil
	}
}

func testAccServiceV1SyslogConfig(name, domain string) string {
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
  syslog {
    name               = "somesyslogname"
    address            = "127.0.0.1"
	response_condition = "response_condition_test"
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceV1SyslogConfig_update(name, domain string) string {
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
  syslog {
    name               = "somesyslogname"
    address            = "127.0.0.1"
    port               = 514
	response_condition = "response_condition_test"
    message_type       = "blank"
  }
  syslog {
    name    = "somesyslogname2"
    address = "127.0.0.2"
    port    = 10514
  }
  force_destroy = true
}`, name, domain)
}

func testAccServiceV1SyslogConfig_formatVersion(name, domain string) string {
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
  syslog {
    name           = "somesyslogname"
    address        = "127.0.0.1"
    port           = 514
    format         = "%%a %%l %%u %%t %%m %%U%%q %%H %%>s %%b %%T"
    format_version = 2
  }
  force_destroy = true
}`, name, domain)
}
