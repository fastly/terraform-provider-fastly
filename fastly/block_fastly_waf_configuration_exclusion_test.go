package fastly

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccFastlyServiceWAFVersionV1FlattenWAFRuleExclusions(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAFRuleExclusion
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.WAFRuleExclusion{
				{
					ID:            "abc",
					Number:        gofastly.Int(1),
					Name:          gofastly.String("index page"),
					ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
					Condition:     gofastly.String("req.url.basename == \"index.html\""),
					Rules: []*gofastly.WAFRule{
						{
							ID:       "2029718",
							ModSecID: 2029718,
						},
						{
							ID:       "1010070",
							ModSecID: 1010070,
						},
					},
				},
				{
					ID:            "def",
					Number:        gofastly.Int(2),
					Name:          gofastly.String("index php"),
					ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
					Condition:     gofastly.String("req.url.basename == \"index.php\""),
					Rules: []*gofastly.WAFRule{
						{
							ID:       "1010070",
							ModSecID: 1010070,
						},
					},
				},
				{
					ID:            "ghi",
					Number:        gofastly.Int(3),
					Name:          gofastly.String("index asp"),
					ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeWAF),
					Condition:     gofastly.String("req.url.basename == \"index.asp\""),
				},
			},
			local: []map[string]interface{}{
				{
					"number":          1,
					"name":            "index page",
					"exclusion_type":  "rule",
					"condition":       "req.url.basename == \"index.html\"",
					"modsec_rule_ids": schema.NewSet(schema.HashInt, []interface{}{2029718, 1010070}),
				},
				{
					"number":          2,
					"name":            "index php",
					"exclusion_type":  "rule",
					"condition":       "req.url.basename == \"index.php\"",
					"modsec_rule_ids": schema.NewSet(schema.HashInt, []interface{}{1010070}),
				},
				{
					"number":         3,
					"name":           "index asp",
					"exclusion_type": "waf",
					"condition":      "req.url.basename == \"index.asp\"",
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFRuleExclusions(c.remote)
		local := c.local
		assertEqualsSliceOfMaps(t, out, local)
	}
}

func TestAccFastlyServiceWAFVersionV1Validation(t *testing.T) {
	// As we use a 'table test' which executes a `resource.Test` multiple times within a for-loop, we don't utilise the
	// `resource.ParallelTest` function but instead call t.Parallel(). The use of t.Parallel() must happen outside of
	// the for-loop otherwise it would be executed multiple times, leading to a runtime panic.
	t.Parallel()

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	cases := []struct {
		exclusions      []gofastly.WAFRuleExclusion
		expectedMessage string
	}{
		{
			exclusions: []gofastly.WAFRuleExclusion{
				{
					Name:          gofastly.String("index page"),
					ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeWAF),
					Condition:     gofastly.String("req.url.basename == \"index.html\""),
					Rules: []*gofastly.WAFRule{
						{
							ModSecID: 2029718,
						},
					},
				},
			},
			expectedMessage: "must not set \"modsec_rule_ids\" with \"waf\" exclusion type in exclusion \"index page\"",
		},
		{
			exclusions: []gofastly.WAFRuleExclusion{
				{
					Name:          gofastly.String("index page"),
					ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
					Condition:     gofastly.String("req.url.basename == \"index.html\""),
				},
			},
			expectedMessage: "must set \"modsec_rule_ids\" with \"rule\" exclusion type in exclusion \"index page\"",
		},
	}

	for _, c := range cases {
		wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20, true)
		exclusionsTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRuleExclusions(c.exclusions)

		wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, "", exclusionsTF1)

		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
			},
			ProviderFactories: testAccProviders,
			CheckDestroy:      testAccCheckServiceVCLDestroy,
			Steps: []resource.TestStep{
				{
					Config:      testAccFastlyServiceWAFVersionV1(name, wafVer1),
					ExpectError: regexp.MustCompile(c.expectedMessage),
				},
			},
		})
	}
}

func TestAccFastlyServiceWAFVersionV1AddUpdateDeleteExclusions(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	rules := []gofastly.WAFActiveRule{
		{
			ModSecID: 21032607,
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
	}

	exclusions1 := []gofastly.WAFRuleExclusion{
		{
			Name:          gofastly.String("index page"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.html\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2029718,
				},
				{
					ModSecID: 2037405,
				},
			},
		},
		{
			Name:          gofastly.String("index php"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.php\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2037405,
				},
			},
		},
		{
			Name:          gofastly.String("index asp"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeWAF),
			Condition:     gofastly.String("req.url.basename == \"index.asp\""),
		},
	}

	exclusions2 := []gofastly.WAFRuleExclusion{
		{
			Name:          gofastly.String("index page"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.html\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 21032607,
				},
			},
		},
		{
			Name:          gofastly.String("index php"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index.php\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2037405,
				},
			},
		},
		{
			Name:          gofastly.String("index new page"),
			ExclusionType: gofastly.String(gofastly.WAFRuleExclusionTypeRule),
			Condition:     gofastly.String("req.url.basename == \"index-new.html\""),
			Rules: []*gofastly.WAFRule{
				{
					ModSecID: 2037405,
				},
			},
		},
	}

	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20, true)
	rulesTF := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules)
	exclusionsTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRuleExclusions(exclusions1)
	exclusionsTF2 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRuleExclusions(exclusions2)

	wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF, exclusionsTF1)
	wafVer2 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF, exclusionsTF2)

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
					testAccCheckFastlyServiceWAFVersionV1CheckExclusions(&service, exclusions1, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceVCLExists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckExclusions(&service, exclusions2, 2),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceWAFVersionV1CheckExclusions(service *gofastly.ServiceDetail, expected []gofastly.WAFRuleExclusion, wafVerNo int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		exclResp, err := conn.ListAllWAFRuleExclusions(&gofastly.ListAllWAFRuleExclusionsInput{
			WAFID:            waf.ID,
			WAFVersionNumber: wafVerNo,
			Include:          []string{"waf_rules"},
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		actual := exclResp.Items
		if len(expected) != len(actual) {
			return fmt.Errorf("Error matching rule exclusions slice sizes :\nexpected: %#v\ngot: %#v", len(expected), len(actual))
		}

		sort.Slice(expected, func(i, j int) bool {
			return *expected[i].Name < *expected[j].Name
		})

		sort.Slice(actual, func(i, j int) bool {
			return *actual[i].Name < *actual[j].Name
		})

		for i, expectedExcl := range expected {
			actualExcl := actual[i]

			if *expectedExcl.Name != *actualExcl.Name {
				return fmt.Errorf("Error matching Name:\nexpected: %#v\ngot: %#v", *expectedExcl.Name, *actualExcl.Name)
			}

			if *expectedExcl.Condition != *actualExcl.Condition {
				return fmt.Errorf("Error matching Condition:\nexpected: %#v\ngot: %#v", *expectedExcl.Condition, *actualExcl.Condition)
			}

			if *expectedExcl.ExclusionType != *actualExcl.ExclusionType {
				return fmt.Errorf("Error matching ExclusionType:\nexpected: %#v\ngot: %#v", *expectedExcl.ExclusionType, *actualExcl.ExclusionType)
			}

			if len(expectedExcl.Rules) != len(actualExcl.Rules) {
				return fmt.Errorf("Error matching rules slice sizes :\nexpected: %#v\ngot: %#v", len(expectedExcl.Rules), len(actualExcl.Rules))
			}

			sort.Slice(expectedExcl.Rules, func(i, j int) bool {
				return expectedExcl.Rules[i].ModSecID < expectedExcl.Rules[j].ModSecID
			})
			sort.Slice(actualExcl.Rules, func(i, j int) bool {
				return actualExcl.Rules[i].ModSecID < actualExcl.Rules[j].ModSecID
			})

			for k, expectedRule := range expectedExcl.Rules {
				actualRule := actualExcl.Rules[k]
				if expectedRule.ModSecID != actualRule.ModSecID {
					return fmt.Errorf("Error matching Rule ModsecId:\nexpected: %#v\ngot: %#v", expectedRule.ModSecID, actualRule.ModSecID)
				}
			}
		}

		return nil
	}
}

func testAccCheckFastlyServiceWAFVersionV1ComposeWAFRuleExclusions(exclusions []gofastly.WAFRuleExclusion) string {
	var result string
	for _, excl := range exclusions {
		var modsecIds []string
		for _, r := range excl.Rules {
			modsecIds = append(modsecIds, strconv.Itoa(r.ModSecID))
		}

		rule := fmt.Sprintf(`
          rule_exclusion {
            name = "%s"
            condition = "%s"
            exclusion_type = "%s"
            modsec_rule_ids = [%s]
          }`, *excl.Name, strings.ReplaceAll(*excl.Condition, "\"", "\\\""), *excl.ExclusionType, strings.Join(modsecIds, ","))
		result = result + rule
	}
	return result
}
