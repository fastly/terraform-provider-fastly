package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenHeaders(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Header
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Header{
				{
					Name:              "myheader",
					Action:            "delete",
					IgnoreIfSet:       true,
					Type:              "cache",
					Destination:       "http.aws-id",
					Source:            "",
					Regex:             "",
					Substitution:      "",
					Priority:          100,
					RequestCondition:  "",
					CacheCondition:    "",
					ResponseCondition: "",
				},
			},
			local: []map[string]any{
				{
					"name":          "myheader",
					"action":        gofastly.HeaderActionDelete,
					"ignore_if_set": true,
					"type":          gofastly.HeaderTypeCache,
					"destination":   "http.aws-id",
					"priority":      int(100),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHeaders(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestFastlyServiceVCL_BuildHeaders(t *testing.T) {
	cases := []struct {
		remote *gofastly.CreateHeaderInput
		local  map[string]any
	}{
		{
			remote: &gofastly.CreateHeaderInput{
				Action:            gofastly.HeaderActionPtr(gofastly.HeaderActionDelete),
				CacheCondition:    gofastly.String("test"),
				Destination:       gofastly.String("http.aws-id"),
				IgnoreIfSet:       gofastly.CBool(true),
				Name:              gofastly.String("someheadder"),
				Priority:          gofastly.Int(100),
				Regex:             gofastly.String("test"),
				RequestCondition:  gofastly.String("test"),
				ResponseCondition: gofastly.String("test"),
				Source:            gofastly.String("test"),
				Substitution:      gofastly.String("test"),
				Type:              gofastly.HeaderTypePtr(gofastly.HeaderTypeCache),
			},
			local: map[string]any{
				"action":             "delete",
				"cache_condition":    "test",
				"destination":        "http.aws-id",
				"ignore_if_set":      true,
				"name":               "someheadder",
				"priority":           100,
				"regex":              "test",
				"request_condition":  "test",
				"response_condition": "test",
				"source":             "test",
				"substitution":       "test",
				"type":               "cache",
			},
		},
		{
			remote: &gofastly.CreateHeaderInput{
				Action:            gofastly.HeaderActionPtr(gofastly.HeaderActionSet),
				CacheCondition:    gofastly.String(""),
				Destination:       gofastly.String("http.aws-id"),
				IgnoreIfSet:       gofastly.CBool(false),
				Name:              gofastly.String("someheadder"),
				Priority:          gofastly.Int(100),
				Regex:             gofastly.String(""),
				RequestCondition:  gofastly.String(""),
				ResponseCondition: gofastly.String(""),
				Source:            gofastly.String("http.server-name"),
				Substitution:      gofastly.String(""),
				Type:              gofastly.HeaderTypePtr(gofastly.HeaderTypeCache),
			},
			local: map[string]any{
				"action":             "set",
				"cache_condition":    "",
				"destination":        "http.aws-id",
				"ignore_if_set":      false,
				"name":               "someheadder",
				"priority":           100,
				"regex":              "",
				"request_condition":  "",
				"response_condition": "",
				"source":             "http.server-name",
				"substitution":       "",
				"type":               "cache",
			},
		},
	}

	for _, c := range cases {
		out, _ := buildHeader(c.local)
		if !reflect.DeepEqual(out, c.remote) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.remote, out)
		}
	}
}

func TestAccFastlyServiceVCL_headers_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Header{
		ServiceVersion: 1,
		Name:           "remove x-amz-request-id",
		Destination:    "http.x-amz-request-id",
		Type:           "cache",
		Action:         "delete",
		Priority:       100,
	}

	log2 := gofastly.Header{
		ServiceVersion: 1,
		Name:           "remove s3 server",
		Destination:    "http.Server",
		Type:           "cache",
		Action:         "delete",
		IgnoreIfSet:    true,
		Priority:       100,
	}

	log3 := gofastly.Header{
		ServiceVersion: 1,
		Name:           "DESTROY S3",
		Destination:    "http.Server",
		Type:           "cache",
		Action:         "delete",
		Priority:       100,
	}

	log4 := gofastly.Header{
		ServiceVersion:    1,
		Name:              "Add server name",
		Destination:       "http.server-name",
		Type:              "request",
		Action:            "set",
		Source:            "server.identity",
		Priority:          100,
		RequestCondition:  "test_req_condition",
		CacheCondition:    "test_cache_condition",
		ResponseCondition: "test_res_condition",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLHeadersConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHeaderAttributes(&service, []*gofastly.Header{&log1, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "header.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLHeadersConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLHeaderAttributes(&service, []*gofastly.Header{&log1, &log3, &log4}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "header.#", "3"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLHeaderAttributes(service *gofastly.ServiceDetail, headers []*gofastly.Header) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		headersList, err := conn.ListHeaders(&gofastly.ListHeadersInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Headers for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(headersList) != len(headers) {
			return fmt.Errorf("healthcheck List count mismatch, expected (%d), got (%d)", len(headers), len(headersList))
		}

		var found int
		for _, h := range headers {
			for _, lh := range headersList {
				if h.Name == lh.Name {
					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ID
					h.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lh.CreatedAt = nil
					lh.UpdatedAt = nil
					if !reflect.DeepEqual(h, lh) {
						return fmt.Errorf("bad match Header match, expected (%#v), got (%#v)", h, lh)
					}
					found++
				}
			}
		}

		if found != len(headers) {
			return fmt.Errorf("error matching Header rules")
		}

		return nil
	}
}

func testAccServiceVCLHeadersConfig(name, domain string) string {
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

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  header {
    destination   = "http.Server"
    type          = "cache"
    action        = "delete"
    name          = "remove s3 server"
    ignore_if_set = "true"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLHeadersConfigUpdate(name, domain string) string {
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

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  header {
    destination = "http.Server"
    type        = "cache"
    action      = "delete"
    name        = "DESTROY S3"
  }

	condition {
    name      = "test_req_condition"
    type      = "REQUEST"
    priority  = 5
    statement = "req.url ~ \"^/foo/bar$\""
  }

	condition {
    name      = "test_cache_condition"
    type      = "CACHE"
    priority  = 9
    statement = "req.url ~ \"^/articles/\""
  }

	condition {
    name      = "test_res_condition"
    type      = "RESPONSE"
    priority  = 10
    statement = "resp.status == 404"
  }

  header {
    destination 			 = "http.server-name"
    type        			 = "request"
    action      			 = "set"
    source      			 = "server.identity"
    name        			 = "Add server name"
		request_condition  = "test_req_condition"
		cache_condition    = "test_cache_condition"
		response_condition = "test_res_condition"
  }

  force_destroy = true
}`, name, domain)
}
