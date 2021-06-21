package fastly

import (
	"context"
	"errors"
	"log"
	"sort"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyWAFRules() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyWAFRulesRead,

		Schema: map[string]*schema.Schema{
			"publishers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of publishers to be used as filters for the data set.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of tags to be used as filters for the data set.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"exclude_modsec_rule_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of modsecurity rules IDs to be excluded from the data set.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of rules that results from any given combination of filters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"modsec_rule_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The modsecurity rule ID.",
						},
						"latest_revision_number": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The modsecurity rule's latest revision.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The modsecurity rule's type.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyWAFRulesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*FastlyClient).conn
	input := &gofastly.ListAllWAFRulesInput{
		Include: "waf_rule_revisions",
	}

	if v, ok := d.GetOk("publishers"); ok {
		l := v.([]interface{})
		for i := range l {
			input.FilterPublishers = append(input.FilterPublishers, l[i].(string))
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		l := v.([]interface{})
		for i := range l {
			input.FilterTagNames = append(input.FilterTagNames, l[i].(string))
		}
	}

	if v, ok := d.GetOk("exclude_modsec_rule_ids"); ok {
		l := v.([]interface{})
		for i := range l {
			input.ExcludeMocSecIDs = append(input.ExcludeMocSecIDs, l[i].(int))
		}
	}

	log.Printf("[INFO] Reading WAF rules with ops: %#v", input)
	res, err := conn.ListAllWAFRules(input)
	if err != nil {
		return diag.Errorf("error listing WAF rules: %s", err)
	}

	rules := flattenWAFRules(res.Items)

	d.SetId(strconv.Itoa(createFiltersHash(input)))
	if err := d.Set("rules", rules); err != nil {
		return diag.Errorf("error setting WAF rules: %s", err)
	}

	return nil
}

func createFiltersHash(i *gofastly.ListAllWAFRulesInput) int {
	var result string
	for _, v := range i.FilterPublishers {
		result = result + v
	}
	for _, v := range i.FilterTagNames {
		result = result + v
	}
	for _, v := range i.ExcludeMocSecIDs {
		result = result + strconv.Itoa(v)
	}
	return hashcode.String(result)
}

func flattenWAFRules(ruleList []*gofastly.WAFRule) []map[string]interface{} {

	rl := make([]map[string]interface{}, len(ruleList))
	if len(ruleList) == 0 {
		return rl
	}

	for i, r := range ruleList {

		latestRevisionNumber := 1
		if latestRevision, err := determineLatestRuleRevision(r.Revisions); err == nil {
			latestRevisionNumber = latestRevision.Revision
		}

		rulesMapString := map[string]interface{}{
			"modsec_rule_id":         r.ModSecID,
			"latest_revision_number": latestRevisionNumber,
			"type":                   r.Type,
		}

		// Prune any empty values that come from the default string value in structs.
		for k, v := range rulesMapString {
			if v == "" {
				delete(rulesMapString, k)
			}
		}
		rl[i] = rulesMapString
	}

	return rl
}

func determineLatestRuleRevision(revisions []*gofastly.WAFRuleRevision) (*gofastly.WAFRuleRevision, error) {

	if len(revisions) == 0 {
		return nil, errors.New("the list of WAFRuleRevisions cannot be empty")
	}

	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Revision > revisions[j].Revision
	})

	return revisions[0], nil
}
