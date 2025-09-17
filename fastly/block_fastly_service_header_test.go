package fastly

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
)

func TestResourceFastlyFlattenHeaders(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Header
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Header{
				{
					Name:              gofastly.ToPointer("myheader"),
					Action:            gofastly.ToPointer(gofastly.HeaderActionDelete),
					IgnoreIfSet:       gofastly.ToPointer(true),
					Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
					Destination:       gofastly.ToPointer("http.aws-id"),
					Source:            gofastly.ToPointer(""),
					Regex:             gofastly.ToPointer(""),
					Substitution:      gofastly.ToPointer(""),
					Priority:          gofastly.ToPointer(100),
					RequestCondition:  gofastly.ToPointer(""),
					CacheCondition:    gofastly.ToPointer(""),
					ResponseCondition: gofastly.ToPointer(""),
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
				Action:            gofastly.ToPointer(gofastly.HeaderActionDelete),
				CacheCondition:    gofastly.ToPointer("test"),
				Destination:       gofastly.ToPointer("http.aws-id"),
				IgnoreIfSet:       gofastly.ToPointer(gofastly.Compatibool(true)),
				Name:              gofastly.ToPointer("someheadder"),
				Priority:          gofastly.ToPointer(100),
				Regex:             gofastly.ToPointer("test"),
				RequestCondition:  gofastly.ToPointer("test"),
				ResponseCondition: gofastly.ToPointer("test"),
				Source:            gofastly.ToPointer("test"),
				Substitution:      gofastly.ToPointer("test"),
				Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
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
				Action:            gofastly.ToPointer(gofastly.HeaderActionSet),
				CacheCondition:    gofastly.ToPointer(""),
				Destination:       gofastly.ToPointer("http.aws-id"),
				IgnoreIfSet:       gofastly.ToPointer(gofastly.Compatibool(false)),
				Name:              gofastly.ToPointer("someheadder"),
				Priority:          gofastly.ToPointer(100),
				Regex:             gofastly.ToPointer(""),
				RequestCondition:  gofastly.ToPointer(""),
				ResponseCondition: gofastly.ToPointer(""),
				Source:            gofastly.ToPointer("http.server-name"),
				Substitution:      gofastly.ToPointer(""),
				Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
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
		out := buildHeader(c.local)
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
		Action:            gofastly.ToPointer(gofastly.HeaderActionDelete),
		CacheCondition:    gofastly.ToPointer(""),
		Destination:       gofastly.ToPointer("http.x-amz-request-id"),
		IgnoreIfSet:       gofastly.ToPointer(false),
		Name:              gofastly.ToPointer("remove x-amz-request-id"),
		Priority:          gofastly.ToPointer(100),
		Regex:             gofastly.ToPointer(""),
		RequestCondition:  gofastly.ToPointer(""),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Source:            gofastly.ToPointer(""),
		Substitution:      gofastly.ToPointer(""),
		Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
	}

	log2 := gofastly.Header{
		Action:            gofastly.ToPointer(gofastly.HeaderActionDelete),
		CacheCondition:    gofastly.ToPointer(""),
		Destination:       gofastly.ToPointer("http.Server"),
		IgnoreIfSet:       gofastly.ToPointer(true),
		Name:              gofastly.ToPointer("remove s3 server"),
		Priority:          gofastly.ToPointer(100),
		Regex:             gofastly.ToPointer(""),
		RequestCondition:  gofastly.ToPointer(""),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Source:            gofastly.ToPointer(""),
		Substitution:      gofastly.ToPointer(""),
		Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
	}

	log3 := gofastly.Header{
		Action:            gofastly.ToPointer(gofastly.HeaderActionDelete),
		CacheCondition:    gofastly.ToPointer(""),
		Destination:       gofastly.ToPointer("http.Server"),
		IgnoreIfSet:       gofastly.ToPointer(false),
		Name:              gofastly.ToPointer("DESTROY S3"),
		Priority:          gofastly.ToPointer(100),
		Regex:             gofastly.ToPointer(""),
		RequestCondition:  gofastly.ToPointer(""),
		ResponseCondition: gofastly.ToPointer(""),
		ServiceVersion:    gofastly.ToPointer(1),
		Source:            gofastly.ToPointer(""),
		Substitution:      gofastly.ToPointer(""),
		Type:              gofastly.ToPointer(gofastly.HeaderTypeCache),
	}

	log4 := gofastly.Header{
		Action:            gofastly.ToPointer(gofastly.HeaderActionSet),
		CacheCondition:    gofastly.ToPointer("test_cache_condition"),
		Destination:       gofastly.ToPointer("http.server-name"),
		IgnoreIfSet:       gofastly.ToPointer(false),
		Name:              gofastly.ToPointer("Add server name"),
		Priority:          gofastly.ToPointer(100),
		Regex:             gofastly.ToPointer(""),
		RequestCondition:  gofastly.ToPointer("test_req_condition"),
		ResponseCondition: gofastly.ToPointer("test_res_condition"),
		ServiceVersion:    gofastly.ToPointer(1),
		Source:            gofastly.ToPointer("server.identity"),
		Substitution:      gofastly.ToPointer(""),
		Type:              gofastly.ToPointer(gofastly.HeaderTypeRequest),
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
		headersList, err := conn.ListHeaders(context.TODO(), &gofastly.ListHeadersInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Headers for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(headersList) != len(headers) {
			return fmt.Errorf("healthcheck List count mismatch, expected (%d), got (%d)", len(headers), len(headersList))
		}

		var found int
		for _, h := range headers {
			for _, lh := range headersList {
				if gofastly.ToValue(h.Name) == gofastly.ToValue(lh.Name) {
					// we don't know these things ahead of time, so populate them now
					h.ServiceID = service.ServiceID
					h.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
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
