package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenDynamicSnippets(t *testing.T) {

	cases := []struct {
		remote []*gofastly.Snippet
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Snippet{
				{
					Name:     "recv_test_01",
					Type:     gofastly.SnippetTypeRecv,
					Priority: 110,
					Content:  "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}",
					Dynamic:  1,
				},
			},
			local: []map[string]interface{}{
				{
					"name":     "recv_test_01",
					"type":     gofastly.SnippetTypeRecv,
					"priority": 110,
				},
			},
		},
		{
			remote: []*gofastly.Snippet{
				{
					Name:     "recv_test_02",
					Type:     gofastly.SnippetTypeRecv,
					Priority: 110,
					Content:  "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}",
					Dynamic:  0,
				},
			},
			local: []map[string]interface{}(nil),
		},
	}

	for _, c := range cases {
		out := flattenDynamicSnippets(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}

}

func TestAccFastlyServiceV1DynamicSnippet_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s1 := gofastly.Snippet{
		Name:     "recv_test",
		Type:     gofastly.SnippetTypeRecv,
		Priority: int(110),
		Dynamic:  1,
	}

	updatedS1 := gofastly.Snippet{
		Name:     "recv_test",
		Type:     gofastly.SnippetTypeRecv,
		Priority: int(110),
		Dynamic:  1,
	}
	updatedS2 := gofastly.Snippet{
		Name:     "fetch_test",
		Type:     gofastly.SnippetTypeFetch,
		Priority: int(50),
		Dynamic:  1,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1DynamicSnippet(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DynamicSnippetAttributes(&service, []*gofastly.Snippet{&s1}),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.name", "recv_test"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.type", "recv"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.priority", "110"),
				),
			},

			{
				Config: testAccServiceV1DynamicSnippet_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1DynamicSnippetAttributes(&service, []*gofastly.Snippet{&updatedS1, &updatedS2}),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.#", "2"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.name", "recv_test"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.type", "recv"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.1074880210.priority", "110"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.3801163681.name", "fetch_test"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.3801163681.type", "fetch"),
					resource.TestCheckResourceAttr("fastly_service_v1.foo", "dynamicsnippet.3801163681.priority", "50"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DynamicSnippetAttributes(service *gofastly.ServiceDetail, snippets []*gofastly.Snippet) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		sList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up VCL Dynamic Snippets for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(sList) != len(snippets) {
			return fmt.Errorf("Dynamic Snippet List count mismatch, expected (%d), got (%d)", len(snippets), len(sList))
		}

		var found int
		for _, expected := range snippets {
			for _, lr := range sList {
				if expected.Name == lr.Name {
					expected.ServiceID = service.ID
					expected.ServiceVersion = service.ActiveVersion.Number

					// We don't know these things ahead of time, so ignore them
					lr.ID = ""
					lr.CreatedAt = nil
					lr.UpdatedAt = nil

					if !reflect.DeepEqual(expected, lr) {
						return fmt.Errorf("Unexpected VCL Dynamic Snippet.\nExpected: %#v\nGot: %#v\n", expected, lr)
					}
					found++
				}
			}
		}

		if found != len(snippets) {
			return fmt.Errorf("Error matching VCL Dynamic Snippets. Found: %d / Expected: %d", found, len(snippets))
		}

		return nil
	}
}

func testAccServiceV1DynamicSnippet(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dynamicsnippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}

func testAccServiceV1DynamicSnippet_update(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dynamicsnippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
  }

  dynamicsnippet {
    name     = "fetch_test"
    type     = "fetch"
    priority = 50
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}
