package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenLoggly(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Loggly
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Loggly{
				{
					Version:       1,
					Name:          "loggly-endpoint",
					Token:         "token",
					FormatVersion: 2,
				},
			},
			local: []map[string]interface{}{
				{
					"name":           "loggly-endpoint",
					"token":          "token",
					"format_version": uint(2),
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

func TestAccFastlyServiceV1_logging_loggly_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	log1 := gofastly.Loggly{
		Version:       1,
		Name:          "loggly-endpoint",
		Token:         "s3cr3t",
		FormatVersion: 2,
		Format:        "%h %l %u %t \"%r\" %>s %b",
	}

	log1_after_update := gofastly.Loggly{
		Version:       1,
		Name:          "loggly-endpoint",
		Token:         "secret",
		FormatVersion: 2,
		Format:        "%h %l %u %t \"%r\" %>s %b %T",
	}

	log2 := gofastly.Loggly{
		Version:       1,
		Name:          "another-loggly-endpoint",
		Token:         "another-token",
		FormatVersion: 2,
		Format:        "%h %l %u %t \"%r\" %>s %b",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1LogglyConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogglyAttributes(&service, []*gofastly.Loggly{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_loggly.#", "1"),
				),
			},

			{
				Config: testAccServiceV1LogglyConfig_update(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1LogglyAttributes(&service, []*gofastly.Loggly{&log1_after_update, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "logging_loggly.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1LogglyAttributes(service *gofastly.ServiceDetail, loggly []*gofastly.Loggly) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		logglyList, err := conn.ListLoggly(&gofastly.ListLogglyInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Loggly Logging for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(logglyList) != len(loggly) {
			return fmt.Errorf("Loggly List count mismatch, expected (%d), got (%d)", len(loggly), len(logglyList))
		}

		log.Printf("[DEBUG] logglyList = %#v\n", logglyList)

		var found int
		for _, e := range loggly {
			for _, el := range logglyList {
				if e.Name == el.Name {
					// we don't know these things ahead of time, so populate them now
					e.ServiceID = service.ID
					e.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					el.CreatedAt = nil
					el.UpdatedAt = nil
					if diff := cmp.Diff(e, el); diff != "" {
						return fmt.Errorf("Bad match Loggly logging match: %s", diff)
					}
					found++
				}
			}
		}

		if found != len(loggly) {
			return fmt.Errorf("Error matching Loggly Logging rules")
		}

		return nil
	}
}

func testAccServiceV1LogglyConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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

func testAccServiceV1LogglyConfig_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
