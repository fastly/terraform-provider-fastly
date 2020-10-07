package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"strconv"
)

var wafExclusion = &schema.Schema{
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
				Description: "The modsec rule ids to exclude.",
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

func readWAFExclusions(meta interface{}, d *schema.ResourceData, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn
	wafID := d.Get("waf_id").(string)

	resp, e := conn.ListAllWAFExclusions(&gofastly.ListAllWAFExclusionsInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Include:          strToPtr("waf_rules"),
	})

	if e != nil {
		return e
	}

	err := d.Set("exclusion", flattenWAFExclusions(resp.Items))

	if err != nil {
		log.Printf("[WARN] Error setting WAF exclusions for (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenWAFExclusions(exclusions []*gofastly.WAFExclusion) []map[string]interface{} {
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

func updateWAFExclusions(d *schema.ResourceData, meta interface{}, wafID string, wafVersionNumber int) error {

	os, ns := d.GetChange("exclusion")

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

	err = deleteWAFExclusion(remove, meta, wafID, wafVersionNumber)
	if err != nil {
		return err
	}

	err = createWAFExclusion(add, meta, wafID, wafVersionNumber)
	if err != nil {
		return err
	}

	return nil
}

func deleteWAFExclusion(remove []interface{}, meta interface{}, wafID string, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn

	for _, aRaw := range remove {
		a := aRaw.(map[string]interface{})

		err := conn.DeleteWAFExclusion(&gofastly.DeleteWAFExclusionInput{
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

func createWAFExclusion(add []interface{}, meta interface{}, wafID string, wafVersionNumber int) error {
	conn := meta.(*FastlyClient).conn

	for _, aRaw := range add {
		a := aRaw.(map[string]interface{})

		var rules []*gofastly.WAFRule
		if a["exclusion_type"] == gofastly.WAFExclusionTypeRule {
			for _, ruleId := range a["modsec_rule_ids"].(*schema.Set).List() {
				rules = append(rules, &gofastly.WAFRule{
					ID: strconv.Itoa(ruleId.(int)),
				})
			}
		} else {
			rules = nil
		}

		_, err := conn.CreateWAFExclusion(&gofastly.CreateWAFExclusionInput{
			WAFID:            wafID,
			WAFVersionNumber: wafVersionNumber,
			WAFExclusion: &gofastly.WAFExclusion{
				Name:          strToPtr(a["name"].(string)),
				ExclusionType: strToPtr(a["exclusion_type"].(string)),
				Condition:     strToPtr(a["condition"].(string)),
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
			gofastly.WAFExclusionTypeRule,
			gofastly.WAFExclusionTypeWAF,
		},
		false,
	)
}

func validateWAFExclusion(d *schema.ResourceDiff) error {
	for _, i := range d.Get("exclusion").(*schema.Set).List() {
		wafExclusion := i.(map[string]interface{})

		if wafExclusion["exclusion_type"] == gofastly.WAFExclusionTypeWAF && len(wafExclusion["modsec_rule_ids"].(*schema.Set).List()) > 0 {
			return fmt.Errorf("must not set \"modsec_rule_ids\" with \"waf\" exclusion type in exclusion \"%s\"", wafExclusion["name"])
		}
		if wafExclusion["exclusion_type"] == gofastly.WAFExclusionTypeRule && len(wafExclusion["modsec_rule_ids"].(*schema.Set).List()) == 0 {
			return fmt.Errorf("must set \"modsec_rule_ids\" with \"rule\" exclusion type in exclusion \"%s\"", wafExclusion["name"])
		}
	}
	return nil
}
