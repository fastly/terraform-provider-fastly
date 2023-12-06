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

func TestResourceFastlyFlattenRequestSettings(t *testing.T) {
	cases := []struct {
		remote []*gofastly.RequestSetting
		local  []map[string]any
	}{
		{
			remote: []*gofastly.RequestSetting{
				{
					Name:             gofastly.ToPointer("alt_backend"),
					RequestCondition: gofastly.ToPointer("serve_alt_backend"),
					DefaultHost:      gofastly.ToPointer("tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"),
					XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),
					MaxStaleAge:      gofastly.ToPointer(90),
					Action:           gofastly.ToPointer(gofastly.RequestSettingActionPass),
				},
			},
			local: []map[string]any{
				{
					"name":              "alt_backend",
					"request_condition": "serve_alt_backend",
					"default_host":      "tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com",
					"xff":               gofastly.RequestSettingXFFAppend,
					"max_stale_age":     90,
					"action":            gofastly.RequestSettingActionPass,
					"bypass_busy_wait":  false,
					"force_miss":        false,
					"force_ssl":         false,
					"geo_headers":       false,
					"timer_support":     false,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenRequestSettings(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceVCLRequestSetting_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	rq1 := gofastly.RequestSetting{
		Name:             gofastly.ToPointer("alt_backend"),
		RequestCondition: gofastly.ToPointer("serve_alt_backend"),
		DefaultHost:      gofastly.ToPointer("tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"),
		XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),
		MaxStaleAge:      gofastly.ToPointer(90),
	}
	rq2 := gofastly.RequestSetting{
		Name:             gofastly.ToPointer("alt_backend"),
		RequestCondition: gofastly.ToPointer("serve_alt_backend"),
		DefaultHost:      gofastly.ToPointer("tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"),
		XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),
		MaxStaleAge:      gofastly.ToPointer(900),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLRequestSetting(name, domainName1, "90"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLRequestSettingsAttributes(&service, []*gofastly.RequestSetting{&rq1}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "request_setting.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "condition.#", "1"),
				),
			},
			{
				Config: testAccServiceVCLRequestSetting(name, domainName1, "900"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLRequestSettingsAttributes(&service, []*gofastly.RequestSetting{&rq2}),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "request_setting.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_vcl.foo", "condition.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLRequestSettingsAttributes(service *gofastly.ServiceDetail, rqs []*gofastly.RequestSetting) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		rqList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
			ServiceID:      gofastly.ToValue(service.ID),
			ServiceVersion: gofastly.ToValue(service.ActiveVersion.Number),
		})
		if err != nil {
			return fmt.Errorf("error looking up Request Setting for (%s), version (%v): %s", gofastly.ToValue(service.Name), gofastly.ToValue(service.ActiveVersion.Number), err)
		}

		if len(rqList) != len(rqs) {
			return fmt.Errorf("request Setting List count mismatch, expected (%d), got (%d)", len(rqs), len(rqList))
		}

		var found int
		for _, r := range rqs {
			for _, lr := range rqList {
				if gofastly.ToValue(r.Name) == gofastly.ToValue(lr.Name) {
					// we don't know these things ahead of time, so populate them now
					r.ServiceID = service.ID
					r.ServiceVersion = service.ActiveVersion.Number
					// We don't track these, so clear them out because we also won't know
					// these ahead of time
					lr.CreatedAt = nil
					lr.UpdatedAt = nil
					if !reflect.DeepEqual(r, lr) {
						return fmt.Errorf("bad match Request Setting match, expected (%#v), got (%#v)", r, lr)
					}
					found++
				}
			}
		}

		if found != len(rqs) {
			return fmt.Errorf("error matching Request Setting rules (%d/%d)", found, len(rqs))
		}

		return nil
	}
}

func testAccServiceVCLRequestSetting(name, domain, maxStaleAge string) string {
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
    type      = "REQUEST"
    priority  = 10
    statement = "req.url ~ \"^/alt/\""
  }

  request_setting {
    default_host      = "tftestingother.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name              = "alt_backend"
    request_condition = "serve_alt_backend"
    max_stale_age     = %s
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}`, name, domain, maxStaleAge)
}
