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

func TestResourceFastlyFlattenGzips(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Gzip
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Gzip{
				{
					Name:       gofastly.ToPointer("somegzip"),
					Extensions: gofastly.ToPointer("css"),
				},
			},
			local: []map[string]any{
				{
					"name":       "somegzip",
					"extensions": []any{"css"},
				},
			},
		},
		{
			remote: []*gofastly.Gzip{
				{
					Name:         gofastly.ToPointer("somegzip"),
					Extensions:   gofastly.ToPointer("css json js"),
					ContentTypes: gofastly.ToPointer("text/html"),
				},
				{
					Name:         gofastly.ToPointer("someothergzip"),
					Extensions:   gofastly.ToPointer("css js"),
					ContentTypes: gofastly.ToPointer("text/html text/xml"),
				},
			},
			local: []map[string]any{
				{
					"name":          "somegzip",
					"extensions":    []any{"css", "json", "js"},
					"content_types": []any{"text/html"},
				},
				{
					"name":          "someothergzip",
					"extensions":    []any{"css", "js"},
					"content_types": []any{"text/html", "text/xml"},
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenGzips(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCL_gzips_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	log1 := gofastly.Gzip{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("gzip file types"),
		Extensions:     gofastly.ToPointer("css js"),
		CacheCondition: gofastly.ToPointer("testing_condition"),
	}

	log2 := gofastly.Gzip{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("gzip extensions"),
		ContentTypes:   gofastly.ToPointer("text/css text/html"),
	}

	log3 := gofastly.Gzip{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("all"),
		Extensions:     gofastly.ToPointer("css js html"),
		ContentTypes:   gofastly.ToPointer("text/javascript application/x-javascript application/javascript text/css text/html"),
	}

	log4 := gofastly.Gzip{
		ServiceVersion: gofastly.ToPointer(1),
		Name:           gofastly.ToPointer("all"),
		Extensions:     gofastly.ToPointer("css"),
		ContentTypes:   gofastly.ToPointer("application/x-javascript text/javascript"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLGzipsConfig(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGzipsAttributes(&service, []*gofastly.Gzip{&log1, &log2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "gzip.#", "2"),
				),
			},

			{
				Config: testAccServiceVCLGzipsConfigDeleteCreate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGzipsAttributes(&service, []*gofastly.Gzip{&log3}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "gzip.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLGzipsConfigUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLGzipsAttributes(&service, []*gofastly.Gzip{&log4}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "gzip.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLGzipsAttributes(service *gofastly.ServiceDetail, gzips []*gofastly.Gzip) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		gzipsList, err := conn.ListGzips(&gofastly.ListGzipsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Gzips for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(gzipsList) != len(gzips) {
			return fmt.Errorf("gzip count mismatch, expected (%d), got (%d)", len(gzips), len(gzipsList))
		}

		var found int
		for _, g := range gzips {
			for _, lg := range gzipsList {
				if gofastly.ToValue(g.Name) == gofastly.ToValue(lg.Name) {
					// we don't know these things ahead of time, so populate them now
					g.ServiceID = service.ID
					g.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lg.CreatedAt = nil
					lg.UpdatedAt = nil
					// If empty value is sent, default value is assigned automatically by the API
					// and so we ignore these fields in response
					if gofastly.ToValue(g.Extensions) == "" {
						lg.Extensions = gofastly.ToPointer("")
					}
					if gofastly.ToValue(g.ContentTypes) == "" {
						lg.ContentTypes = gofastly.ToPointer("")
					}
					if !reflect.DeepEqual(g, lg) {
						return fmt.Errorf("bad match Gzip match, expected (%#v), got (%#v)", g, lg)
					}
					found++
				}
			}
		}

		if found != len(gzips) {
			return fmt.Errorf("error matching Gzip rules")
		}

		return nil
	}
}

func testAccServiceVCLGzipsConfig(name, domain string) string {
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
    name      = "testing_condition"
    type      = "CACHE"
    priority  = 10
    statement = "req.url ~ \"^/articles/\""
  }

  gzip {
    name       			= "gzip file types"
    extensions 			= ["css", "js"]
		cache_condition = "testing_condition"
  }

  gzip {
    name          = "gzip extensions"
    content_types = ["text/css", "text/html"]
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLGzipsConfigDeleteCreate(name, domain string) string {
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

  gzip {
    name       = "all"
    extensions = ["css", "js", "html"]

    content_types = [
	  "text/javascript",
	  "application/x-javascript",
	  "application/javascript",
	  "text/css",
	  "text/html",
    ]
  }

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLGzipsConfigUpdate(name, domain string) string {
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

  gzip {
    name       = "all"
    extensions = ["css"]

    content_types = [
      "application/x-javascript",
      "text/javascript",
    ]
  }

  force_destroy = true
}`, name, domain)
}
