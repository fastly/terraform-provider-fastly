package fastly

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceFastlyFlattenSplunk(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Splunk
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Splunk{
				{
					Name:              "test-splunk",
					URL:               "https://mysplunkendpoint.example.com/services/collector/event",
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     1,
					ResponseCondition: "error_response",
					Placement:         "waf_debug",
					Token:             "test-token",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "test-splunk",
					"url":                "https://mysplunkendpoint.example.com/services/collector/event",
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(1),
					"response_condition": "error_response",
					"placement":          "waf_debug",
					"token":              "test-token",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSplunks(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1_splunk_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLogOne := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %t \"%r\" %>s %b",
		FormatVersion:     1,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	splunkLogOneUpdated := gofastly.Splunk{
		Name:              "test-splunk-1",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "error_response_5XX",
	}

	splunkLogTwo := gofastly.Splunk{
		Name:              "test-splunk-2",
		URL:               "https://mysplunkendpoint.example.com/services/collector/event",
		Token:             "test-token",
		Format:            "%h %l %u %{now}V %{req.method}V %{req.url}V %>s %{resp.http.Content-Length}V",
		FormatVersion:     2,
		Placement:         "waf_debug",
		ResponseCondition: "ok_response_2XX",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_complete(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOne}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},

			{
				Config: testAccServiceV1SplunkConfig_update(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLogOneUpdated, &splunkLogTwo}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "2"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_default(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	splunkLog := gofastly.Splunk{
		Name:          "test-splunk",
		URL:           "https://mysplunkendpoint.example.com/services/collector/event",
		Token:         "test-token",
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_default(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLog}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_splunk_env(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	// set env variable to something we expect
	resetEnv := setSplunkEnv("test-token", t)
	defer resetEnv()

	splunkLog := gofastly.Splunk{
		Name:          "test-splunk",
		URL:           "https://mysplunkendpoint.example.com/services/collector/event",
		Token:         "test-token",
		Format:        "%h %l %u %t \"%r\" %>s %b",
		FormatVersion: 2,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1SplunkConfig_env(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SplunkAttributes(&service, []*gofastly.Splunk{&splunkLog}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", serviceName),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "splunk.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1SplunkAttributes(service *gofastly.ServiceDetail, localSplunkList []*gofastly.Splunk) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		remoteSplunkList, err := conn.ListSplunks(&gofastly.ListSplunksInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Splunk for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(remoteSplunkList) != len(localSplunkList) {
			return fmt.Errorf("Splunk List count mismatch, expected (%d), got (%d)", len(localSplunkList), len(remoteSplunkList))
		}

		var found int
		for _, ls := range localSplunkList {
			for _, rs := range remoteSplunkList {
				if ls.Name == rs.Name {
					// we don't know these things ahead of time, so populate them now
					ls.ServiceID = service.ID
					ls.Version = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					rs.CreatedAt = nil
					rs.UpdatedAt = nil
					if !reflect.DeepEqual(ls, rs) {
						return fmt.Errorf("Bad match Splunk logging match, expected (%#v), got (%#v)", ls, rs)
					}
					found++
				}
			}
		}

		if found != len(localSplunkList) {
			return fmt.Errorf("Error matching Splunk rules")
		}

		return nil
	}
}

func testAccServiceV1SplunkConfig_complete(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = "%%h %%l %%u %%t \"%%r\" %%>s %%b"
    format_version     = 1
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1SplunkConfig_update(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  condition {
    name      = "error_response_5XX"
    statement = "resp.status >= 500 && resp.status < 600"
    priority  = 10
    type      = "RESPONSE"
  }

  condition {
    name      = "ok_response_2XX"
    statement = "resp.status >= 200 && resp.status < 300"
    priority  = 10
    type      = "RESPONSE"
  }

  splunk {
    name               = "test-splunk-1"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = "%%h %%l %%u %%{now}V %%{req.method}V %%{req.url}V %%>s %%{resp.http.Content-Length}V"
    format_version     = 2
    placement          = "waf_debug"
    response_condition = "error_response_5XX"
  }

  splunk {
    name               = "test-splunk-2"
    url                = "https://mysplunkendpoint.example.com/services/collector/event"
    token              = "test-token"
    format             = "%%h %%l %%u %%{now}V %%{req.method}V %%{req.url}V %%>s %%{resp.http.Content-Length}V"
    format_version     = 2
    placement          = "waf_debug"
    response_condition = "ok_response_2XX"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1SplunkConfig_default(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  splunk {
    name  = "test-splunk"
    url   = "https://mysplunkendpoint.example.com/services/collector/event"
    token = "test-token"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func testAccServiceV1SplunkConfig_env(serviceName string) string {
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "tf-test-backend"
  }

  splunk {
    name = "test-splunk"
    url  = "https://mysplunkendpoint.example.com/services/collector/event"
  }

  force_destroy = true
}`, serviceName, domainName)
}

func setSplunkEnv(token string, t *testing.T) func() {
	e := getSplunkEnv()
	// Set all the envs to a dummy value
	if err := os.Setenv("FASTLY_SPLUNK_TOKEN", token); err != nil {
		t.Fatalf("Error setting env var FASTLY_SPLUNK_TOKEN: %s", err)
	}

	return func() {
		// re-set all the envs we unset above
		if err := os.Setenv("FASTLY_SPLUNK_TOKEN", e.Token); err != nil {
			t.Fatalf("Error resetting env var FASTLY_SPLUNK_TOKEN: %s", err)
		}
	}
}

// struct to preserve the current environment
type currentSplunkEnv struct {
	Token string
}

func getSplunkEnv() *currentSplunkEnv {
	// Grab the existing Fastly Splunk token and preserve, in the off chance
	// they're actually set in the enviornment
	return &currentSplunkEnv{
		Token: os.Getenv("FASTLY_SPLUNK_TOKEN"),
	}
}
