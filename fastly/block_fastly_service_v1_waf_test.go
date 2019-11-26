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
				},
			},
			local: []map[string]interface{}{
				{
					"waf_id":             "test1",
					"prefetch_condition": "prefetch",
					"response_object":    "response",
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
	waf := composeWAF(condition, response)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFAddAndRemove(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	waf := composeWAF(condition, response)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", waf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, "", ""),
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
	waf := composeWAF(condition, response)
	updatedWaf := composeWAF(condition, updateResponse)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, extraResponse, updatedWaf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, updateResponse, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFUpdateCondition(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updatedCondition := "UpdatedPrefetch"
	waf := composeWAF(condition, response)
	updatedWaf := composeWAF(updatedCondition, response)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAF(name, "", waf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1WAF(name, extraCondition, updatedWaf),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, updatedCondition),
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

func testAccCheckFastlyServiceV1AttributesWAF(service *gofastly.ServiceDetail, name, response, condition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

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
			return fmt.Errorf("WAF response mismatch, expected: %s, got: %#v", response, resp.Items[0].Response)
		}

		if resp.Items[0].PrefetchCondition != condition {
			return fmt.Errorf("WAF condition mismatch, expected: %#v, got: %#v", condition, resp.Items[0].PrefetchCondition)
		}

		return nil
	}
}

func composeWAF(condition, response string) string {
	return fmt.Sprintf(`
		waf { 
			prefetch_condition = "%s" 
			response_object = "%s"
		}`, condition, response)
}

func testAccServiceV1WAF(name, extraHCL, waf string) string {

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
}`, name, domainName, backendName, extraHCL, waf)

}
