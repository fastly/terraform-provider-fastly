package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
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
					Action:           gofastly.ToPointer(gofastly.RequestSettingActionPass),
					DefaultHost:      gofastly.ToPointer("http-me.glitch.me"),
					MaxStaleAge:      gofastly.ToPointer(90),
					Name:             gofastly.ToPointer("alt_backend"),
					RequestCondition: gofastly.ToPointer("serve_alt_backend"),
					XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),
				},
			},
			local: []map[string]any{
				{
					"action": gofastly.RequestSettingActionPass,
					// "bypass_busy_wait":  false,
					"default_host": "http-me.glitch.me",
					// "force_miss":        false,
					// "force_ssl":         false,
					// "geo_headers":       false,
					"max_stale_age":     90,
					"name":              "alt_backend",
					"request_condition": "serve_alt_backend",
					// "timer_support":     false,
					"xff": gofastly.RequestSettingXFFAppend,
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
		DefaultHost:      gofastly.ToPointer("http-me.glitch.me"),
		MaxStaleAge:      gofastly.ToPointer(90),
		Name:             gofastly.ToPointer("alt_backend"),
		RequestCondition: gofastly.ToPointer("serve_alt_backend"),
		XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),

		// We only set a few attributes in our TF config (see above).
		// For all the other attributes (with the exception of `action` and `xff`,
		// which are only sent to the API if they have a non-zero string value)
		// the default value for their types are still sent to the API
		// and so the API responds with those default values. Hence we have to set
		// those defaults below...
		BypassBusyWait: gofastly.ToPointer(false),
		ForceMiss:      gofastly.ToPointer(false),
		ForceSSL:       gofastly.ToPointer(false),
		GeoHeaders:     gofastly.ToPointer(false),
		HashKeys:       gofastly.ToPointer(""),
		TimerSupport:   gofastly.ToPointer(false),
	}

	rq2 := gofastly.RequestSetting{
		Action:           gofastly.ToPointer(gofastly.RequestSettingActionLookup),
		DefaultHost:      gofastly.ToPointer("http-me.glitch.me"),
		MaxStaleAge:      gofastly.ToPointer(900),
		Name:             gofastly.ToPointer("alt_backend"),
		RequestCondition: gofastly.ToPointer("serve_alt_backend"),
		XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),

		// We only set a few attributes in our TF config (see above).
		// For all the other attributes (with the exception of `action` and `xff`,
		// which are only sent to the API if they have a non-zero string value)
		// the default value for their types are still sent to the API
		// and so the API responds with those default values. Hence we have to set
		// those defaults below...
		BypassBusyWait: gofastly.ToPointer(false),
		ForceMiss:      gofastly.ToPointer(false),
		ForceSSL:       gofastly.ToPointer(false),
		GeoHeaders:     gofastly.ToPointer(false),
		HashKeys:       gofastly.ToPointer(""),
		TimerSupport:   gofastly.ToPointer(false),
	}
	rq3 := gofastly.RequestSetting{
		Action:           gofastly.ToPointer(gofastly.RequestSettingActionUnset),
		DefaultHost:      gofastly.ToPointer("http-me.glitch.me"),
		MaxStaleAge:      gofastly.ToPointer(900),
		Name:             gofastly.ToPointer("alt_backend"),
		RequestCondition: gofastly.ToPointer("serve_alt_backend"),
		XForwardedFor:    gofastly.ToPointer(gofastly.RequestSettingXFFAppend),

		// We only set a few attributes in our TF config (see above).
		// For all the other attributes (with the exception of `action` and `xff`,
		// which are only sent to the API if they have a non-zero string value)
		// the default value for their types are still sent to the API
		// and so the API responds with those default values. Hence we have to set
		// those defaults below...
		BypassBusyWait: gofastly.ToPointer(false),
		ForceMiss:      gofastly.ToPointer(false),
		ForceSSL:       gofastly.ToPointer(false),
		GeoHeaders:     gofastly.ToPointer(false),
		HashKeys:       gofastly.ToPointer(""),
		TimerSupport:   gofastly.ToPointer(false),
	}

	createAction := ""        // initially we expect no action to be set in HTTP request.
	updateAction1 := "lookup" // give it a value and expect it to be set.
	updateAction2 := ""       // set an empty value and expect the empty string to be sent to the API.

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceVCLRequestSetting(name, domainName1, createAction, "90"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLRequestSettingsAttributes(&service, []*gofastly.RequestSetting{&rq1}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "request_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "condition.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "request_setting.*", map[string]string{
						"action":        "", // IMPORTANT: To validate this attribute we need at least one map key to have a non-empty value (hence the `max_stale_age` check below).
						"max_stale_age": "900",
					}),
				),
			},
			{
				Config: testAccServiceVCLRequestSetting(name, domainName1, updateAction1, "900"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLRequestSettingsAttributes(&service, []*gofastly.RequestSetting{&rq2}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "request_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "condition.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "request_setting.*", map[string]string{
						"action": "lookup",
					}),
				),
			},
			{
				Config: testAccServiceVCLRequestSetting(name, domainName1, updateAction2, "900"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists("fastly_service_vcl.foo", &service),
					testAccCheckFastlyServiceVCLRequestSettingsAttributes(&service, []*gofastly.RequestSetting{&rq3}),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "name", name),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "request_setting.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_vcl.foo", "condition.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("fastly_service_vcl.foo", "request_setting.*", map[string]string{
						"action":        "", // IMPORTANT: To validate this attribute we need at least one map key to have a non-empty value (hence the `max_stale_age` check below).
						"max_stale_age": "900",
					}),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceVCLRequestSettingsAttributes(service *gofastly.ServiceDetail, rqs []*gofastly.RequestSetting) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		rqList, err := conn.ListRequestSettings(&gofastly.ListRequestSettingsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
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
					r.ServiceID = service.ServiceID
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

func testAccServiceVCLRequestSetting(name, domain, action, maxStaleAge string) string {
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

  backend {
    address = "server-timing-test.glitch.me"
    name    = "Other Glitch Test Site"
    port    = 80
  }

  condition {
    name      = "serve_alt_backend"
    type      = "REQUEST"
    priority  = 10
    statement = "req.url ~ \"^/alt/\""
  }

  request_setting {
    action            = "%s"
    default_host      = "http-me.glitch.me"
    name              = "alt_backend"
    request_condition = "serve_alt_backend"
    max_stale_age     = %s
  }

  default_host = "http-me.glitch.me"

  force_destroy = true
}`, name, domain, action, maxStaleAge)
}
