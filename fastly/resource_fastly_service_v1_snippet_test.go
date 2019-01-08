package fastly

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func TestAccFastlyServiceV1Snippet_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s1 := gofastly.Snippet{
		Name:     "recv_test",
		Type:     gofastly.SnippetTypeRecv,
		Priority: int(110),
		Content:  "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}",
	}

	s2 := gofastly.Snippet{
		Name:     "fetch_test",
		Type:     gofastly.SnippetTypeFetch,
		Priority: int(50),
		Content:  "if ( req.url ~ \"^/pass\" ) {\n return(pass);\n}",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Snippet(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SnippetAttributes(&service, []*gofastly.Snippet{&s1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "vcl_snippet.#", "1"),
				),
			},

			{
				Config: testAccServiceV1Snippet_update(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1SnippetAttributes(&service, []*gofastly.Snippet{&s1, &s2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "vcl_snippet.#", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "vcl_snippet.2.name", "fetch_test"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1SnippetAttributes(service *gofastly.ServiceDetail, snippets []*gofastly.Snippet) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		sList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up VCL Snippets for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(sList) != len(snippets) {
			return fmt.Errorf("Snippet List count mismatch, expected (%d), got (%d)", len(snippets), len(sList))
		}

		var found int
		for _, r := range snippets {
			for _, lr := range sList {
				if r.Name == lr.Name {
					// we don't know these things ahead of time, so populate them now
					r.ServiceID = service.ID
					r.Version = service.ActiveVersion.Number
					if !reflect.DeepEqual(r, lr) {
						return fmt.Errorf("Bad VCL Snippet match, expected (%#v), got (%#v)", r, lr)
					}
					found++
				}
			}
		}

		if found != len(snippets) {
			return fmt.Errorf("Error matching VCL Snippets (%d/%d)", found, len(snippets))
		}

		return nil
	}
}

func testAccServiceV1Snippet(name, domain string) string {
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

  vcl_snippet {
    name = "recv_test"
    type = "recv"
    priority = 110
    content = "myfunc()"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}

func testAccServiceV1Snippet_update(name, domain string) string {
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

  vcl_snippet {
    name = "recv_test"
    type = "recv"
    priority = 110
    content = "if (resp.status >= 500) { restart }"
  }

  vcl_snippet {
    name = "fetch_test"
    type = "fetch"
    priority = 50
    content = "restart"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}
