package fastly

import (
	"fmt"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var wafRuleExclusion = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the exclusion.",
			},
			"condition": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A conditional expression in VCL used to determine if the condition is met.",
			},
			"exclusion_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of exclusion.",
				ValidateFunc: validateExecutionType(),
			},
			"modsec_rule_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The modsec rule IDs to exclude.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "A sequential ID assigned to the exclusion.",
			},
		},
	},
}

func readWAFRuleExclusions(meta interface{}, d *schema.ResourceData, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn
	wafID := d.Get("waf_id").(string)

	resp, e := conn.ListAllWAFRuleExclusions(&gofastly.ListAllWAFRuleExclusionsInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Include:          []string{"waf_rules"},
	})

	if e != nil {
		return e
	}

	err := d.Set("rule_exclusion", flattenWAFRuleExclusions(resp.Items))

	if err != nil {
		log.Printf("[WARN] Error setting WAF rule exclusions for (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenWAFRuleExclusions(exclusions []*gofastly.WAFRuleExclusion) []map[string]interface{} {
	var result []map[string]interface{}

	for _, exclusion := range exclusions {

		m := make(map[string]interface{})
		if exclusion.Name != nil {
			m["name"] = *exclusion.Name
		}
		if exclusion.Number != nil {
			m["number"] = *exclusion.Number
		}
		if exclusion.Condition != nil {
			m["condition"] = *exclusion.Condition
		}
		if exclusion.ExclusionType != nil {
			m["exclusion_type"] = *exclusion.ExclusionType
		}

		var rules []interface{}
		for _, rule := range exclusion.Rules {
			rules = append(rules, rule.ModSecID)
		}
		if len(rules) > 0 {
			m["modsec_rule_ids"] = schema.NewSet(schema.HashInt, rules)
		}
		result = append(result, m)
	}

	return result
}

func updateWAFRuleExclusions(d *schema.ResourceData, meta interface{}, wafID string, wafVersionNumber int) error {

	os, ns := d.GetChange("rule_exclusion")

	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)

	add := nss.Difference(oss).List()
	remove := oss.Difference(nss).List()

	var err error

	err = deleteWAFRuleExclusion(remove, meta, wafID, wafVersionNumber)
	if err != nil {
		return err
	}

	err = createWAFRuleExclusion(add, meta, wafID, wafVersionNumber)
	if err != nil {
		return err
	}

	return nil
}

func deleteWAFRuleExclusion(remove []interface{}, meta interface{}, wafID string, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn

	for _, aRaw := range remove {
		a := aRaw.(map[string]interface{})

		err := conn.DeleteWAFRuleExclusion(&gofastly.DeleteWAFRuleExclusionInput{
			Number:           a["number"].(int),
			WAFID:            wafID,
			WAFVersionNumber: wafVersionNumber,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func createWAFRuleExclusion(add []interface{}, meta interface{}, wafID string, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn

	for _, aRaw := range add {
		a := aRaw.(map[string]interface{})

		var rules []*gofastly.WAFRule
		if a["exclusion_type"] == gofastly.WAFRuleExclusionTypeRule {
			for _, ruleId := range a["modsec_rule_ids"].(*schema.Set).List() {
				rules = append(rules, &gofastly.WAFRule{
					ID: strconv.Itoa(ruleId.(int)),
				})
			}
		} else {
			rules = nil
		}

		_, err := conn.CreateWAFRuleExclusion(&gofastly.CreateWAFRuleExclusionInput{
			WAFID:            wafID,
			WAFVersionNumber: wafVersionNumber,
			WAFRuleExclusion: &gofastly.WAFRuleExclusion{
				Name:          gofastly.String(a["name"].(string)),
				ExclusionType: gofastly.String(a["exclusion_type"].(string)),
				Condition:     gofastly.String(a["condition"].(string)),
				Rules:         rules,
			},
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func validateExecutionType() schema.SchemaValidateFunc {
	return validation.StringInSlice(
		[]string{
			gofastly.WAFRuleExclusionTypeRule,
			gofastly.WAFRuleExclusionTypeWAF,
		},
		false,
	)
}

func validateWAFRuleExclusion(d *schema.ResourceDiff) error {
	for _, i := range d.Get("rule_exclusion").(*schema.Set).List() {
		wafRuleExclusion := i.(map[string]interface{})

		if wafRuleExclusion["exclusion_type"] == gofastly.WAFRuleExclusionTypeWAF && len(wafRuleExclusion["modsec_rule_ids"].(*schema.Set).List()) > 0 {
			return fmt.Errorf("must not set \"modsec_rule_ids\" with \"waf\" exclusion type in exclusion \"%s\"", wafRuleExclusion["name"])
		}
		if wafRuleExclusion["exclusion_type"] == gofastly.WAFRuleExclusionTypeRule && len(wafRuleExclusion["modsec_rule_ids"].(*schema.Set).List()) == 0 {
			return fmt.Errorf("must set \"modsec_rule_ids\" with \"rule\" exclusion type in exclusion \"%s\"", wafRuleExclusion["name"])
		}
	}
	return nil
}
