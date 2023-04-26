package fastly

import (
	"fmt"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var wafRuleExclusion = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"condition": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A conditional expression in VCL used to determine if the condition is met",
			},
			"exclusion_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The type of rule exclusion. Values are `rule` to exclude the specified rule(s), or `waf` to disable the Web Application Firewall",
				ValidateDiagFunc: validateExecutionType(),
			},
			"modsec_rule_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Set of modsecurity IDs to be excluded. No rules should be provided when `exclusion_type` is `waf`. The rules need to be configured on the Web Application Firewall to be excluded",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of rule exclusion",
			},
			"number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The numeric ID assigned to the WAF Rule Exclusion",
			},
		},
	},
}

func readWAFRuleExclusions(meta any, d *schema.ResourceData, wafVersionNumber int) error {
	conn := meta.(*APIClient).conn
	wafID := d.Get("waf_id").(string)

	remoteState, e := conn.ListAllWAFRuleExclusions(&gofastly.ListAllWAFRuleExclusionsInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Include:          []string{"waf_rules"},
	})

	if e != nil {
		return e
	}

	err := d.Set("rule_exclusion", flattenWAFRuleExclusions(remoteState.Items))
	if err != nil {
		log.Printf("[WARN] Error setting WAF rule exclusions for (%s): %s", d.Id(), err)
	}

	return nil
}

// flattenWAFRuleExclusions models data into format suitable for saving to Terraform state.
func flattenWAFRuleExclusions(remoteState []*gofastly.WAFRuleExclusion) []map[string]any {
	var result []map[string]any

	for _, resource := range remoteState {
		data := make(map[string]any)
		if resource.Name != nil {
			data["name"] = *resource.Name
		}
		if resource.Number != nil {
			data["number"] = *resource.Number
		}
		if resource.Condition != nil {
			data["condition"] = *resource.Condition
		}
		if resource.ExclusionType != nil {
			data["exclusion_type"] = *resource.ExclusionType
		}

		var rules []any
		for _, rule := range resource.Rules {
			rules = append(rules, rule.ModSecID)
		}
		if len(rules) > 0 {
			data["modsec_rule_ids"] = schema.NewSet(schema.HashInt, rules)
		}
		result = append(result, data)
	}

	return result
}

func updateWAFRuleExclusions(d *schema.ResourceData, meta any, wafID string, wafVersionNumber int) error {
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

func deleteWAFRuleExclusion(remove []any, meta any, wafID string, wafVersionNumber int) error {
	conn := meta.(*APIClient).conn

	for _, aRaw := range remove {
		a := aRaw.(map[string]any)

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

func createWAFRuleExclusion(add []any, meta any, wafID string, wafVersionNumber int) error {
	conn := meta.(*APIClient).conn

	for _, aRaw := range add {
		a := aRaw.(map[string]any)

		var rules []*gofastly.WAFRule
		if a["exclusion_type"] == gofastly.WAFRuleExclusionTypeRule {
			for _, ruleID := range a["modsec_rule_ids"].(*schema.Set).List() {
				rules = append(rules, &gofastly.WAFRule{
					ID: strconv.Itoa(ruleID.(int)),
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

func validateExecutionType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice(
		[]string{
			gofastly.WAFRuleExclusionTypeRule,
			gofastly.WAFRuleExclusionTypeWAF,
		},
		false,
	))
}

func validateWAFRuleExclusion(d *schema.ResourceDiff) error {
	for _, i := range d.Get("rule_exclusion").(*schema.Set).List() {
		wafRuleExclusion := i.(map[string]any)

		if wafRuleExclusion["exclusion_type"] == gofastly.WAFRuleExclusionTypeWAF && len(wafRuleExclusion["modsec_rule_ids"].(*schema.Set).List()) > 0 {
			return fmt.Errorf("must not set \"modsec_rule_ids\" with \"waf\" exclusion type in exclusion \"%s\"", wafRuleExclusion["name"])
		}
		if wafRuleExclusion["exclusion_type"] == gofastly.WAFRuleExclusionTypeRule && len(wafRuleExclusion["modsec_rule_ids"].(*schema.Set).List()) == 0 {
			return fmt.Errorf("must set \"modsec_rule_ids\" with \"rule\" exclusion type in exclusion \"%s\"", wafRuleExclusion["name"])
		}
	}
	return nil
}
