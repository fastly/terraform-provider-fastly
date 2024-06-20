package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenDynamicSnippets(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Snippet
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Snippet{
				{
					Name:     gofastly.ToPointer("recv_test_01"),
					Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
					Priority: gofastly.ToPointer(110),
					Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"),
					Dynamic:  gofastly.ToPointer(1),
				},
			},
			local: []map[string]any{
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
					Name:     gofastly.ToPointer("recv_test_02"),
					Type:     gofastly.ToPointer(gofastly.SnippetTypeRecv),
					Priority: gofastly.ToPointer(110),
					Content:  gofastly.ToPointer("if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"),
					Dynamic:  gofastly.ToPointer(0),
				},
			},
			local: []map[string]any(nil),
		},
	}

	for _, c := range cases {
		out := flattenDynamicSnippets(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCLDynamicSnippet_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	// The `id` field in the API response is a non-deterministic UUID and so in
	// `testAccCheckFastlyServiceVCLDynamicSnippetAttributes()` we have to set
	// it to a pointer to an empty string (hence that's what we also set below).
	s1 := gofastly.Snippet{
		Dynamic:   gofastly.ToPointer(1),
		SnippetID: gofastly.ToPointer(""),
		Name:      gofastly.ToPointer("recv_test"),
		Priority:  gofastly.ToPointer(int(110)),
		Type:      gofastly.ToPointer(gofastly.SnippetTypeRecv),
	}
	updatedS1 := gofastly.Snippet{
		Dynamic:   gofastly.ToPointer(1),
		SnippetID: gofastly.ToPointer(""),
		Name:      gofastly.ToPointer("recv_test"),
		Priority:  gofastly.ToPointer(int(110)),
		Type:      gofastly.ToPointer(gofastly.SnippetTypeRecv),
	}
	updatedS2 := gofastly.Snippet{
		Dynamic:   gofastly.ToPointer(1),
		SnippetID: gofastly.ToPointer(""),
		Name:      gofastly.ToPointer("fetch_test"),
		Priority:  gofastly.ToPointer(50),
		Type:      gofastly.ToPointer(gofastly.SnippetTypeFetch),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLDynamicSnippet(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDynamicSnippetAttributes(&service, []*gofastly.Snippet{&s1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "dynamicsnippet.*", map[string]string{
						"name":     "recv_test",
						"type":     "recv",
						"priority": "110",
					}),
				),
			},

			{
				Config: testAccServiceVCLDynamicSnippetUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLDynamicSnippetAttributes(&service, []*gofastly.Snippet{&updatedS1, &updatedS2}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "dynamicsnippet.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "dynamicsnippet.*", map[string]string{
						"name":     "recv_test",
						"type":     "recv",
						"priority": "110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "dynamicsnippet.*", map[string]string{
						"name":     "fetch_test",
						"type":     "fetch",
						"priority": "50",
					}),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLDynamicSnippetAttributes(service *gofastly.ServiceDetail, snippets []*gofastly.Snippet) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		sList, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up VCL Dynamic Snippets for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(sList) != len(snippets) {
			return fmt.Errorf("dynamic Snippet List count mismatch, expected (%d), got (%d)", len(snippets), len(sList))
		}

		var found int
		for _, expected := range snippets {
			for _, lr := range sList {
				if gofastly.ToValue(expected.Name) == gofastly.ToValue(lr.Name) {
					expected.ServiceID = service.ServiceID
					expected.ServiceVersion = service.ActiveVersion.Number

					// We don't know these things ahead of time, so ignore them
					lr.SnippetID = gofastly.ToPointer("")
					lr.CreatedAt = nil
					lr.UpdatedAt = nil

					if !reflect.DeepEqual(expected, lr) {
						return fmt.Errorf("unexpected VCL Dynamic Snippet (expected: %#v, got: %#v)", expected, lr)
					}
					found++
				}
			}
		}

		if found != len(snippets) {
			return fmt.Errorf("error matching VCL Dynamic Snippets (expected: %d, got: %d)", len(snippets), found)
		}

		return nil
	}
}

func testAccServiceVCLDynamicSnippet(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address = "http-me.glitch.me"
    name    = "Glitch Test Site"
    port    = 80
  }

  dynamicsnippet {
    name     = "recv_test"
    type     = "recv"
    priority = 110
  }

  default_host = "http-me.glitch.me"

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLDynamicSnippetUpdate(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_vcl" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "demo"
  }

  backend {
    address = "http-me.glitch.me"
    name    = "Glitch Test Site"
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

  default_host = "http-me.glitch.me"

  force_destroy = true
}`, name, domain)
}
