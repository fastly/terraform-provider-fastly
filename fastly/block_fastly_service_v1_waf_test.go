package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
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
	request_condition = "ALWAYS_FALSE"
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
	waf := composeWAF(condition, response)

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
	waf := composeWAF(condition, response)
	updatedWaf := composeWAF(condition, updateResponse)

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
	waf := composeWAF(condition, response)
	updatedWaf := composeWAF(updatedCondition, response)

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

func composeWAF(condition, response string) string {
	return fmt.Sprintf(`
		waf { 
			prefetch_condition = "%s" 
			response_object    = "%s"
		}`, condition, response)
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

	# The WAF was updated to insert an ALWAYS_FALSE default condition, which 
	# broke our tests because the terraform state was unaware of the default 
	# condition resource that was being dynamically created by the API. This 
	# meant terraform would flag the difference in state as unexpected, and
	# subsequently produce an error.
	#
	# To resolve this error we define the default condition in our terraform which 
	# prevented the API from creating it, but there was a bug in the API 
	# implementation which meant the name of the condition had to match exactly
	# otherwise it would consider the condition missing.
	#
	# TODO(integralist):
	# Once the bug in the API has been fixed, come back and update the tests so 
	# that we can validate the test terraform code no longer requires the
	# condition name to be ALWAYS_FALSE (e.g. set the name to "foobar").
	#
	# NOTE:
	# If the WAF isn't in place and without that ALWAYS_FALSE condition, the WAF 
	# response object (403) will be served for all inbound traffic. This 
	# condition was added by the WAF team to prevent Fastly from returning a 403 
	# on all of the customer traffic before WAF is provisioned to the service.

	condition {
		name      = "ALWAYS_FALSE"
		priority  = 10
		statement = "!req.url"
		type      = "REQUEST"
	}

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
	request_condition = "ALWAYS_FALSE"
  }

  %s

  force_destroy = true
}
%s
`, name, domainName, backendName, extraHCL, waf, config)

}
