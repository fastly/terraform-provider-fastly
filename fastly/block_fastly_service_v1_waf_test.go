package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var serviceRef = "fastly_service_v1.foo"
var condition = "prefetch"
var response = "response"
var extraResponse = `
  response_object {
	name = "UpdatedResponse"
	status = "403"
	response = "Forbidden"
	content = "content"
  }`

var extraCondition = `
  condition {
	name = "UpdatedPrefetch"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }`

func TestResourceFastlyFlattenWAF(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAF
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.WAF{
				{
					ID:                "test1",
					PrefetchCondition: "prefetch",
					Response:          "response",
					Disabled:          true,
				},
			},
			local: []map[string]interface{}{
				{
					"waf_id":             "test1",
					"prefetch_condition": "prefetch",
					"response_object":    "response",
					"disabled":           true,
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1WAFAdd(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	waf := composeWAF(condition, response, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, response, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFAddAndRemove(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	waf := composeWAF(condition, response, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", waf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1DeletedWAF(&service),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFUpdateResponse(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updateResponse := "UpdatedResponse"
	waf := composeWAF(condition, response, false)
	updatedWaf := composeWAF(condition, updateResponse, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, extraResponse, updatedWaf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, updateResponse, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFUpdateCondition(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updatedCondition := "UpdatedPrefetch"
	waf := composeWAF(condition, response, false)
	updatedWaf := composeWAF(updatedCondition, response, false)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, extraCondition, updatedWaf, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, response, updatedCondition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFDisableEnable(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafEnabled := composeWAF(condition, response, false)
	wafDisabled := composeWAF(condition, response, true)

	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	rulesTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules([]gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "block",
			Revision: 1,
		},
	})
	wafConfig := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF1)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", wafEnabled, wafConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1DisableEnableWAF(&service, false),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", wafDisabled, wafConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1DisableEnableWAF(&service, true),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", wafEnabled, wafConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1DisableEnableWAF(&service, false),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DeletedWAF(service *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		resp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return err
		}

		if len(resp.Items) > 0 {
			return fmt.Errorf("[ERR] Error WAF %s should not be present for (%s), version (%v): %s", resp.Items[0].ID, service.ID, service.ActiveVersion.Number, err)
		}
		return nil
	}
}

func testAccCheckFastlyServiceV1AttributesWAF(service *gofastly.ServiceDetail, response, condition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		resp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(resp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(resp.Items))
		}

		if resp.Items[0].Response != response {
			return fmt.Errorf("[ERR] WAF response mismatch, expected: %s, got: %#v", response, resp.Items[0].Response)
		}

		if resp.Items[0].PrefetchCondition != condition {
			return fmt.Errorf("[ERR] WAF condition mismatch, expected: %#v, got: %#v", condition, resp.Items[0].PrefetchCondition)
		}

		return nil
	}
}

func testAccCheckFastlyServiceV1DisableEnableWAF(service *gofastly.ServiceDetail, disable bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		resp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(resp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(resp.Items))
		}

		if resp.Items[0].Disabled != disable {
			return fmt.Errorf("[ERR] WAF disabled mismatch, expected: %v, got: %v", disable, resp.Items[0].Disabled)
		}
		return nil
	}
}

func composeWAF(condition, response string, disabled bool) string {
	return fmt.Sprintf(`
		waf { 
			prefetch_condition = "%s" 
			response_object    = "%s"
            disabled           = %v 
		}`, condition, response, disabled)
}

func testAccServiceV1WAF(name, extraHCL, waf, config string) string {

	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  %s

  condition {
	name = "prefetch"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }

  response_object {
	name = "response"
	status = "403"
	response = "Forbidden"
	content = "content"
  }

  %s

  force_destroy = true
}
%s
`, name, domainName, backendName, extraHCL, waf, config)

}
