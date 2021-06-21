package fastly

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFastlyWAFRulesDetermineRevision(t *testing.T) {

	cases := []struct {
		remote  []*gofastly.WAFRuleRevision
		local   int
		Errored bool
	}{
		{
			remote:  []*gofastly.WAFRuleRevision{},
			local:   0,
			Errored: true,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 1},
			},
			local:   1,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 1},
				{Revision: 2},
			},
			local:   2,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 3},
				{Revision: 2},
				{Revision: 1},
			},
			local:   3,
			Errored: false,
		},
	}

	for _, c := range cases {
		out, err := determineLatestRuleRevision(c.remote)
		if (err == nil) == c.Errored {
			t.Fatalf("Error expected to be %v but wasn't", c.Errored)
		}
		if out == nil {
			continue
		}
		if c.local != out.Revision {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestFastlyWAFRulesFlattenWAFRules(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAFRule
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.WAFRule{
				{
					ModSecID: 11110000,
					Type:     "type",
					Revisions: []*gofastly.WAFRuleRevision{
						{Revision: 1},
					},
				},
			},
			local: []map[string]interface{}{
				{
					"modsec_rule_id":         11110000,
					"type":                   "type",
					"latest_revision_number": 1,
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFRules(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyWAFRulesPublisherFilter(t *testing.T) {

	wafrulesHCL := `
    publishers = ["owasp"]
    `
	wafrulesHCL2 := `
    publishers = ["owasp","fastly"]
    `
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyWAFRulesCheckByPublisherFilter([]string{"owasp"}),
				),
			},
			{
				Config: testAccFastlyWAFRules(wafrulesHCL2),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyWAFRulesCheckByPublisherFilter([]string{"owasp", "fastly"}),
				),
			},
		},
	})
}

func TestAccFastlyWAFRulesExcludeFilter(t *testing.T) {

	wafrulesHCL := `
    publishers = ["owasp"]
    exclude_modsec_rule_ids = [1010020]
    `
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyWAFRulesCheckByExcludeFilter([]string{"owasp"}, []int{1010020}),
				),
			},
		},
	})
}

func TestAccFastlyWAFRulesTagFilter(t *testing.T) {

	wafrulesHCL := `
    tags = ["CVE-2018-17384"]
    `
	wafrulesHCL2 := `
    tags = ["CVE-2018-17384", "attack-rce"]
    `
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyWAFRulesCheckByTagFilter([]string{"CVE-2018-17384"}),
				),
			},
			{
				Config: testAccFastlyWAFRules(wafrulesHCL2),
				Check: resource.ComposeTestCheckFunc(
					testAccFastlyWAFRulesCheckByTagFilter([]string{"CVE-2018-17384", "attack-rce"}),
				),
			},
		},
	})
}

func testAccFastlyWAFRulesCheckByPublisherFilter(publishers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF rule records: error  %s", err)
		}

		return testAccFastlyWAFRulesCheckAgainstState(s, rulesResp.Items)
	}
}

func testAccFastlyWAFRulesCheckByExcludeFilter(publishers []string, exclusions []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
			ExcludeMocSecIDs: exclusions,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF rule records: error  %s", err)
		}

		return testAccFastlyWAFRulesCheckAgainstState(s, rulesResp.Items)
	}
}

func testAccFastlyWAFRulesCheckByTagFilter(tags []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterTagNames: tags,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF rule records: error  %s", err)
		}

		return testAccFastlyWAFRulesCheckAgainstState(s, rulesResp.Items)
	}
}

func testAccFastlyWAFRulesCheckAgainstState(s *terraform.State, rules []*gofastly.WAFRule) error {
	r := s.RootModule().Resources["data.fastly_waf_rules.r1"]
	a := r.Primary.Attributes

	rulesListSize, err := strconv.Atoi(a["rules.#"])
	if err != nil {
		return err
	}

	if rulesListSize != len(rules) {
		return fmt.Errorf("[ERR] Expected WAF rule size (%d), got (%d)", rulesListSize, len(rules))
	}

	modSecIDs := make(map[string]bool, rulesListSize)
	for i := 0; i < rulesListSize; i++ {
		path := fmt.Sprintf("rules.%d.modsec_rule_id", i)
		modSecIDs[a[path]] = true
	}

	for _, r := range rules {
		if _, ok := modSecIDs[strconv.Itoa(r.ModSecID)]; !ok {
			return fmt.Errorf("[ERR] ModSecurity rule id (%d) not found", r.ModSecID)
		}
	}
	return nil
}

func testAccFastlyWAFRules(filtersHCL string) string {

	return fmt.Sprintf(`
    data "fastly_waf_rules" "r1" {
    %s
    }`, filtersHCL)
}
