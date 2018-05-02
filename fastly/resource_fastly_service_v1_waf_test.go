package fastly

import (
	"flag"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

// TODO As of 11/16/2017 the WAF is enabled by Fastly and can't yet be created from the API. The wafStartingState
// below and all marked TODO items can change when it is possible to create WAF from the API.
var (
	wafTestEnabled   = flag.Bool("WAF", false, "Specify the -WAF flag to run the waf test.")
	wafStartingState = `waf {
				prefetch_condition = "WAF_Conditional"
				response = "custom response"
			}`
)

func TestAccFastlyServiceV1_WAF_basic(t *testing.T) {
	if !*wafTestEnabled {
		t.Skip("WAF can't be enabled from the API and so requires a specific account for the tests. Add the flag 'WAF' to run.")
	}
	var service gofastly.ServiceDetail
	// TODO update when WAF can be created by the API
	// name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	// domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	name := "tf-test-d4o12b40fb"
	domainName1 := "fastly-test.tf-kzvp4rnh3n.com"
	existingID := "5FAGaklrkyrXpJxhal0S8L"

	tests := []struct {
		description  string
		wafTerraform string
		waf          *gofastly.WAF
		owasp        *gofastly.OWASP
	}{
		/* TODO update when WAF can be created by the API
		{
			description:  "No WAF defined",
			wafTerraform: "",
			waf:          nil,
			owasp:        nil,
		},
		{
			description: "Create WAF with undefined OWASP",
			wafTerraform: `waf {
				prefetch_condition = "WAF_Conditional"
				response = "custom response"
			}`,
			waf:   &gofastly.WAF{PrefetchCondition: "WAF_Conditional", Response: "custom response"},
			owasp: nil,
		},
		*/
		{
			description: "WAF with default OWASP, update WAF",
			wafTerraform: `waf {
				prefetch_condition = "WAF_Conditional"
				response = "WAF_Response"
			}`,
			waf:   &gofastly.WAF{PrefetchCondition: "WAF_Conditional", Response: "WAF_Response"},
			owasp: nil,
		},
		{
			description: "WAF with OWASP",
			wafTerraform: `waf {
				prefetch_condition = "WAF_Conditional"
				response = "WAF_Response"
				owasp = {
					AllowedMethods = "GET"
				}
			}`,
			waf:   &gofastly.WAF{PrefetchCondition: "WAF_Conditional", Response: "WAF_Response"},
			owasp: &gofastly.OWASP{AllowedMethods: "GET"},
		},
		{
			description: "Updated OWASP",
			wafTerraform: `waf {
				prefetch_condition = "WAF_Conditional"
				response = "WAF_Response"
				owasp = {
					AllowedMethods = "GET"
					AllowedHTTPVersions = "2"
				}
			}`,
			waf:   &gofastly.WAF{PrefetchCondition: "WAF_Conditional", Response: "WAF_Response"},
			owasp: &gofastly.OWASP{AllowedMethods: "GET", AllowedHTTPVersions: "2"},
		},
		{ // TODO update when WAF can be created by the API
			description:  "Reset to starting state",
			wafTerraform: wafStartingState,
			waf:          &gofastly.WAF{PrefetchCondition: "WAF_Conditional", Response: "custom response"},
			owasp:        &gofastly.OWASP{AllowedMethods: "GET", AllowedHTTPVersions: "2"},
		},
		/* TODO update when WAF can be created by the API
		{
			description:  "No WAF defined, expecting delete",
			wafTerraform: "",
			waf:          nil,
			owasp:        nil,
		},
		*/
	}

	testCase := resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
	}
	for _, test := range tests {
		step := resource.TestStep{
			Config: testAccServiceV1WAFConfig(name, domainName1, test.wafTerraform),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
				testAccCheckFastlyServiceV1WAFAttributes(&service, test.waf, test.owasp),
				resource.TestCheckResourceAttr("fastly_service_v1.foo", "name", name),
			),
			// TODO update when WAF can be created by the API, the import and resource name values are not likely needed
			ImportState:   true,
			ImportStateId: existingID,
			ResourceName:  "fastly_service_v1.foo",
		}
		testCase.Steps = append(testCase.Steps, step)
	}
	resource.Test(t, testCase)
}

func testAccCheckFastlyServiceV1WAFAttributes(service *gofastly.ServiceDetail, want *gofastly.WAF, wantOWASP *gofastly.OWASP) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafs, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAFs for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		switch {
		case want == nil && len(wafs) == 0:
			return nil
		case want == nil && len(wafs) != 0:
			return fmt.Errorf("got WAF %+v, want none", wafs)
		case len(wafs) != 1:
			return fmt.Errorf("got %d WAFs, want 1", len(wafs))
		case wafs[0].PrefetchCondition != want.PrefetchCondition:
			return fmt.Errorf("got PrefetchCondition %q, want %q", wafs[0].PrefetchCondition, want.PrefetchCondition)
		case wafs[0].Response != want.Response:
			return fmt.Errorf("got Response %q, want %q", wafs[0].Response, want.Response)
		}

		owasp, err := conn.GetOWASP(&gofastly.GetOWASPInput{
			Service: service.ID,
			ID:      wafs[0].ID,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up OWASP for (%s), WAF ID (%s): %v", service.Name, wafs[0].ID, err)
		}

		switch {
		case wantOWASP == nil && owasp == nil:
			return nil
		case wantOWASP == nil && owasp != nil:
			return fmt.Errorf("got OWASP %+v, want none", owasp)
		case !reflect.DeepEqual(owasp, wantOWASP):
			return fmt.Errorf("got OWASP %+v, want %+v", owasp, wantOWASP)
		}

		return nil
	}
}

func testAccServiceV1WAFConfig(name, domain string, wafDetails string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  %s

  force_destroy = true
}`, name, domain, wafDetails)
}
