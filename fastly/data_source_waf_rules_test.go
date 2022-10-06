package fastly

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestFastlyWAFRules_DetermineRevision(t *testing.T) {
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

func TestFastlyWAFRules_FlattenWAFRules(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAFRule
		local  []map[string]any
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
			local: []map[string]any{
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

func TestAccFastlyWAFRules_PublisherFilter(t *testing.T) {
	wafrulesHCL := `
    publishers = ["owasp"]
    `
	wafrulesHCL2 := `
    publishers = ["owasp","fastly"]
    `
	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
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

func TestAccFastlyWAFRules_ModSecIDsFilter(t *testing.T) {
	wafrulesHCL := `
    modsec_rule_ids = [1010060, 1010070]
    `

	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(wafrulesHCL),
				Check:  resource.TestCheckResourceAttr("data.fastly_waf_rules.r1", "rules.#", "2"),
			},
		},
	})
}

func TestAccFastlyWAFRules_ExcludeFilter(t *testing.T) {
	wafrulesHCL := `
    publishers = ["owasp"]
    exclude_modsec_rule_ids = [1010020]
    `
	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
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

func TestAccFastlyWAFRules_TagFilter(t *testing.T) {
	wafrulesHCL := `
    tags = ["CVE-2018-17384"]
    `
	wafrulesHCL2 := `
    tags = ["CVE-2018-17384", "attack-rce"]
    `
	// lintignore:XAT001
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceVCLDestroy,
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
		conn := testAccProvider.Meta().(*APIClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAF rule records: error  %s", err)
		}

		return testAccFastlyWAFRulesCheckAgainstState(s, rulesResp.Items)
	}
}

func testAccFastlyWAFRulesCheckByExcludeFilter(publishers []string, exclusions []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
			ExcludeMocSecIDs: exclusions,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAF rule records: error  %s", err)
		}

		return testAccFastlyWAFRulesCheckAgainstState(s, rulesResp.Items)
	}
}

func testAccFastlyWAFRulesCheckByTagFilter(tags []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*APIClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterTagNames: tags,
		})
		if err != nil {
			return fmt.Errorf("error looking up WAF rule records: error  %s", err)
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
		return fmt.Errorf("expected WAF rule size (%d), got (%d)", rulesListSize, len(rules))
	}

	modSecIDs := make(map[string]bool, rulesListSize)
	for i := 0; i < rulesListSize; i++ {
		path := fmt.Sprintf("rules.%d.modsec_rule_id", i)
		modSecIDs[a[path]] = true
	}

	for _, r := range rules {
		if _, ok := modSecIDs[strconv.Itoa(r.ModSecID)]; !ok {
			return fmt.Errorf("the ModSecurity rule id (%d) not found", r.ModSecID)
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
