package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceFastlyFlattenCacheSettings(t *testing.T) {
	cases := []struct {
		remote []*gofastly.CacheSetting
		local  []map[string]any
	}{
		{
			remote: []*gofastly.CacheSetting{
				{
					Name:           "alt_backend",
					Action:         gofastly.CacheSettingActionPass,
					StaleTTL:       3600,
					CacheCondition: "serve_alt_backend",
					TTL:            300,
				},
			},
			local: []map[string]any{
				{
					"name":            "alt_backend",
					"action":          gofastly.CacheSettingActionPass,
					"cache_condition": "serve_alt_backend",
					"stale_ttl":       uint(3600),
					"ttl":             uint(300),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenCacheSettings(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCLCacheSetting_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	cq1 := gofastly.CacheSetting{
		Name:           "alt_backend",
		Action:         "pass",
		StaleTTL:       uint(3600),
		CacheCondition: "serve_alt_backend",
	}

	cq2 := gofastly.CacheSetting{
		Name:           "cache_backend",
		Action:         "restart",
		StaleTTL:       uint(1600),
		CacheCondition: "cache_alt_backend",
		TTL:            uint(300),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLCacheSetting(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLCacheSettingsAttributes(&service, []*gofastly.CacheSetting{&cq1}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "cache_setting.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "condition.#", "1"),
				),
			},

			{
				Config: testAccServiceVCLCacheSettingUpdate(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLCacheSettingsAttributes(&service, []*gofastly.CacheSetting{&cq1, &cq2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "cache_setting.#", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "condition.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLCacheSettingsAttributes(service *gofastly.ServiceDetail, cs []*gofastly.CacheSetting) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		cList, err := conn.ListCacheSettings(&gofastly.ListCacheSettingsInput{
			ServiceID:      service.ID,
			ServiceVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up Cache Setting for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(cList) != len(cs) {
			return fmt.Errorf("cache Setting List count mismatch, expected (%d), got (%d)", len(cs), len(cList))
		}

		var found int
		for _, c := range cs {
			for _, lc := range cList {
				if c.Name == lc.Name {
					// we don't know these things ahead of time, so populate them now
					c.ServiceID = service.ID
					c.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also wont know
					// these ahead of time
					lc.CreatedAt = nil
					lc.UpdatedAt = nil
					if !reflect.DeepEqual(c, lc) {
						return fmt.Errorf("bad match Cache Setting match, expected (%#v), got (%#v)", c, lc)
					}
					found++
				}
			}
		}

		if found != len(cs) {
			return fmt.Errorf("error matching Cache Setting rules (%d/%d)", found, len(cs))
		}

		return nil
	}
}

func testAccServiceVCLCacheSetting(name, domain string) string {
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

  backend {
    address = "tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "OtherAWSS3hosting"
    port    = 80
  }

  condition {
    name      = "serve_alt_backend"
    type      = "CACHE"
    priority  = 10
    statement = "req.url ~ \"^/alt/\""
  }

  cache_setting {
    name            = "alt_backend"
    stale_ttl       = 3600
    cache_condition = "serve_alt_backend"
    action          = "pass"
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}

func testAccServiceVCLCacheSettingUpdate(name, domain string) string {
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

  backend {
    address = "tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "OtherAWSS3hosting"
    port    = 80
  }

  condition {
    name      = "serve_alt_backend"
    type      = "CACHE"
    priority  = 10
    statement = "req.url ~ \"^/alt/\""
  }

  condition {
    name      = "cache_alt_backend"
    type      = "CACHE"
    priority  = 20
    statement = "req.url ~ \"^/cache/\""
  }

  cache_setting {
    name            = "alt_backend"
    stale_ttl       = 3600
    cache_condition = "serve_alt_backend"
    action          = "pass"
  }

  cache_setting {
    name            = "cache_backend"
    stale_ttl       = 1600
    cache_condition = "cache_alt_backend"
    action          = "restart"
    ttl             = 300
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain)
}
