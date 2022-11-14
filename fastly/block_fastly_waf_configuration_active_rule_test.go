package fastly

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceWAFVersionV1_FlattenWAFActiveRules(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAFActiveRule
		local  []map[string]any
	}{
		{
			remote: []*gofastly.WAFActiveRule{
				{
					ModSecID: 1110111,
					Revision: 1,
					Status:   "log",
				},
			},
			local: []map[string]any{
				{
					"modsec_rule_id": 1110111,
					"revision":       1,
					"status":         "log",
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFActiveRules(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceWAFVersionV1_AddUpdateDeleteRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	rules1 := []gofastly.WAFActiveRule{
		{
			ModSecID: 1010090,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2029718,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 910100,
			Status:   "score",
			Revision: 1,
		},
	}
	rules2 := []gofastly.WAFActiveRule{
		// update status
		{
			ModSecID: 1010080,
			Status:   "block",
			Revision: 1,
		},
		{
			ModSecID: 2029718,
			Status:   "block",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "block",
			Revision: 1,
		},
		// update revision
		{
			ModSecID: 910100,
			Status:   "score",
			Revision: 2,
		},
	}
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20, true)
	rulesTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules1)
	wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF1, "")

	rulesTF2 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules2)
	wafVer2 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF2, "")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules1, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules2, 2),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceWAFVersionV1CheckRules(service *gofastly.ServiceDetail, expected []gofastly.WAFActiveRule, wafVerNo int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		ruleResp, err := conn.ListWAFActiveRules(&gofastly.ListWAFActiveRulesInput{
			WAFID:            waf.ID,
			WAFVersionNumber: wafVerNo,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		actual := ruleResp.Items
		if len(expected) != len(actual) {
			return fmt.Errorf("error matching rules slice sizes :\nexpected: %#v\ngot: %#v", len(expected), len(actual))
		}

		sort.Slice(expected[:], func(i, j int) bool {
			return expected[i].ModSecID < expected[j].ModSecID
		})
		sort.Slice(actual[:], func(i, j int) bool {
			return actual[i].ModSecID < actual[j].ModSecID
		})
		for i := range expected {
			if expected[i].ModSecID != actual[i].ModSecID {
				return fmt.Errorf("error matching:\nexpected: %#v\ngot: %#v", expected[i].ModSecID, actual[i].ModSecID)
			}
			if expected[i].Status != actual[i].Status {
				return fmt.Errorf("error matching:\nexpected: %#v\ngot: %#v", expected[i].Status, actual[i].Status)
			}
			if expected[i].Revision != actual[i].Revision {
				return fmt.Errorf("error matching:\nexpected: %#v\ngot: %#v", expected[i].Revision, actual[i].Revision)
			}
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules []gofastly.WAFActiveRule) string {
	var result string
	for _, r := range rules {
		rule := fmt.Sprintf(`
          rule {
            modsec_rule_id = %d
            revision = %d
            status = "%s"
          }`, r.ModSecID, r.Revision, r.Status)
		result = result + rule
	}
	return result
}
