package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenSnippets(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Snippet
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Snippet{
				{
					Name:     gofastly.ToPointer("recv_test"),
					Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
					Priority: gofastly.ToPointer(110),
					Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"),
				},
			},
			local: []map[string]any{
				{
					"name":     "recv_test",
					"type":     gofastly.SnippetTypeRecv,
					"priority": 110,
					"content":  "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}",
				},
			},
		},
		{
			remote: []*gofastly.Snippet{
				{
					Name:     gofastly.ToPointer("recv_test"),
					Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
					Priority: gofastly.ToPointer(110),
					Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"),
					Dynamic:  gofastly.ToPointer(1),
				},
			},
			local: []map[string]any(nil),
		},
	}

	for _, c := range cases {
		out := flattenSnippets(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCLSnippet_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	s1 := gofastly.Snippet{
		Name:     gofastly.ToPointer("recv_test"),
		Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
		Priority: gofastly.ToPointer(110),
		Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"),
	}

	updatedS1 := gofastly.Snippet{
		Name:     gofastly.ToPointer("recv_test"),
		Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
		Priority: gofastly.ToPointer(110),
		Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.different-header = \"true\";\n}"),
	}
	updatedS2 := gofastly.Snippet{
		Name:     gofastly.ToPointer("fetch_test"),
		Type:     gofastly.ToPointer(gofastly.SnippetTypeFetch),
		Priority: gofastly.ToPointer(50),
		Content:  gofastly.ToPointer("restart;\n"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLSnippet(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSnippetAttributes(&service, []*gofastly.Snippet{&s1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "snippet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "snippet.*", map[string]string{
						"name":     "recv_test",
						"type":     "recv",
						"priority": "110",
					}),
				),
			},

			{
				Config: testAccServiceVCLSnippetUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLSnippetAttributes(&service, []*gofastly.Snippet{&updatedS1, &updatedS2}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "snippet.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "snippet.*", map[string]string{
						"name":     "recv_test",
						"type":     "recv",
						"priority": "110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "snippet.*", map[string]string{
						"name":     "fetch_test",
						"type":     "fetch",
						"priority": "50",
					}),
				),
			},

			{
				Config:      testAccServiceVCLMultipleSnippets(name, domainName1),
				ExpectError: regexp.MustCompile("multiple snippets with the same name"),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLSnippetAttributes(service *gofastly.ServiceDetail, snippets []*gofastly.Snippet) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		sList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL Snippets for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(sList) != len(snippets) {
			return fmt.Errorf("snippet List count mismatch, expected (%d), got (%d)", len(snippets), len(sList))
		}

		var found int
		for _, expected := range snippets {
			for _, lr := range sList {
				if gofastly.ToValue(expected.Name) == gofastly.ToValue(lr.Name) {
					expected.ServiceID = service.ID
					expected.ServiceVersion = service.ActiveVersion.Number

					// We don't know these things ahead of time, so ignore them
					lr.ID = gofastly.ToPointer("")
					lr.CreatedAt = nil
					lr.UpdatedAt = nil

					if !reflect.DeepEqual(expected, lr) {
						return fmt.Errorf("unexpected VCL Snippet (expected: %#v, got: %#v)", expected, lr)
					}
					found++
				}
			}
		}

		if found != len(snippets) {
			return fmt.Errorf("error matching VCL Snippets (expected: %d, got: %d)", len(snippets), found)
		}

		return nil
	}
}

func testAccServiceVCLSnippet(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  snippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
    content  = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLSnippetUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  snippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
    content  = "if ( req.url ) {\n set req.http.different-header = \"true\";\n}"
  }

  snippet {
    name     = "fetch_test"
    type     = "fetch"
    priority = 50
    content  = "restart;\n"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}

// IMPORTANT: The following config defines two snippets with the same 'name'.
// Although allowed by the Fastly API, this isn't ideal.
// That's because it results in 'last entry wins' behaviour.
// The Fastly Terraform provider should return an error when generating the diff.
func testAccServiceVCLMultipleSnippets(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
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

  snippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
    content  = "if ( req.url ) {\n set req.http.different-header = \"true\";\n}"
  }

  snippet {
    name     = "fetch_test"
    type     = "fetch"
    priority = 50
    content  = "restart;\n"
  }

  snippet {
    name     = "fetch_test"
    type     = "fetch"
    priority = 110
    content  = "restart;\n"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}
