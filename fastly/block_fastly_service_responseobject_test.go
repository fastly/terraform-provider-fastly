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

func TestResourceFastlyFlattenResponseObjects(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ResponseObject
		local  []map[string]any
	}{
		{
			remote: []*gofastly.ResponseObject{
				{
					ServiceVersion:   gofastly.ToPointer(1),
					Name:             gofastly.ToPointer("responseObjecttesting"),
					Status:           gofastly.ToPointer(200),
					Response:         gofastly.ToPointer("OK"),
					Content:          gofastly.ToPointer("test content"),
					ContentType:      gofastly.ToPointer("text/html"),
					RequestCondition: gofastly.ToPointer("test-request-condition"),
					CacheCondition:   gofastly.ToPointer("test-cache-condition"),
				},
			},
			local: []map[string]any{
				{
					"name":              "responseObjecttesting",
					"status":            200,
					"response":          "OK",
					"content":           "test content",
					"content_type":      "text/html",
					"request_condition": "test-request-condition",
					"cache_condition":   "test-cache-condition",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenResponseObjects(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_response_object_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.ResponseObject{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("responseObjecttesting"),
		Status:           gofastly.ToPointer(200),
		Response:         gofastly.ToPointer("OK"),
		Content:          gofastly.ToPointer("test content"),
		ContentType:      gofastly.ToPointer("text/html"),
		RequestCondition: gofastly.ToPointer("test-request-condition"),
		CacheCondition:   gofastly.ToPointer("test-cache-condition"),
	}

	log2 := gofastly.ResponseObject{
		ServiceVersion:   gofastly.ToPointer(1),
		Name:             gofastly.ToPointer("responseObjecttesting2"),
		Status:           gofastly.ToPointer(404),
		Response:         gofastly.ToPointer("Not Found"),
		Content:          gofastly.ToPointer("some, other, content"),
		ContentType:      gofastly.ToPointer("text/csv"),
		RequestCondition: gofastly.ToPointer("another-test-request-condition"),
		CacheCondition:   gofastly.ToPointer("another-test-cache-condition"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLResponseObjectConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLResponseObjectAttributes(&service, []*gofastly.ResponseObject{&log1}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "response_object.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLResponseObjectConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLResponseObjectAttributes(&service, []*gofastly.ResponseObject{&log1, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "response_object.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLResponseObjectAttributes(service *gofastly.ServiceDetail, responseObjects []*gofastly.ResponseObject) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		responseObjectList, err := conn.ListResponseObjects(&gofastly.ListResponseObjectsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Response Object for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(responseObjectList) != len(responseObjects) {
			return fmt.Errorf("response Object List count mismatch, expected (%d), got (%d)", len(responseObjects), len(responseObjectList))
		}

		var found int
		for _, p := range responseObjects {
			for _, lp := range responseObjectList {
				if gofastly.ToValue(p.Name) == gofastly.ToValue(lp.Name) {
					// we don't know these things ahead of time, so populate them now
					p.ServiceID = service.ID
					p.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					lp.CreatedAt = nil
					lp.UpdatedAt = nil
					if !reflect.DeepEqual(p, lp) {
						return fmt.Errorf("bad match Response Object match, expected (%#v), got (%#v)", p, lp)
					}
					found++
				}
			}
		}

		if found != len(responseObjects) {
			return fmt.Errorf("error matching Response Object rules")
		}

		return nil
	}
}

func testAccServiceVCLResponseObjectConfig(name, domain string) string {
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
    name      = "test-request-condition"
    type      = "REQUEST"
    priority  = 5
    statement = "req.url ~ \"^/foo/bar$\""
  }

	condition {
    name      = "test-cache-condition"
    type      = "CACHE"
    priority  = 9
    statement = "req.url ~ \"^/articles/\""
  }

  response_object {
		name              = "responseObjecttesting"
		status            = 200
		response          = "OK"
		content           = "test content"
		content_type      = "text/html"
		request_condition = "test-request-condition"
		cache_condition   = "test-cache-condition"
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLResponseObjectConfigUpdate(name, domain string) string {
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
    name      = "test-cache-condition"
    type      = "CACHE"
    priority  = 9
    statement = "req.url ~ \"^/articles/\""
  }

	condition {
    name      = "another-test-cache-condition"
    type      = "CACHE"
    priority  = 7
    statement = "req.url ~ \"^/stories/\""
  }

	condition {
    name      = "test-request-condition"
    type      = "REQUEST"
    priority  = 5
    statement = "req.url ~ \"^/foo/bar$\""
  }

	condition {
    name      = "another-test-request-condition"
    type      = "REQUEST"
    priority  = 10
    statement = "req.url ~ \"^/articles$\""
  }

  response_object {
		name              = "responseObjecttesting"
		status            = 200
		response          = "OK"
		content           = "test content"
		content_type      = "text/html"
		request_condition = "test-request-condition"
		cache_condition   = "test-cache-condition"
  }

  response_object {
		name              = "responseObjecttesting2"
		status            = 404
		response          = "Not Found"
		content           = "some, other, content"
		content_type      = "text/csv"
		request_condition = "another-test-request-condition"
		cache_condition   = "another-test-cache-condition"
  }

  force_destroy = true
}`, name, domain)
}
