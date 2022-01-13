package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var activeRule = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The Web Application Firewall rule's status. Allowed values are (`log`, `block` and `score`)",
				ValidateDiagFunc: validateRuleStatusType(),
			},
			"modsec_rule_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Web Application Firewall rule's modsecurity ID",
			},
			"revision": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Web Application Firewall rule's revision. The latest revision will be used if this is not provided",
			},
		},
	},
}

func updateRules(d *schema.ResourceData, meta interface{}, wafID string, Number int) error {

	conn := meta.(*FastlyClient).conn
	os, ns := d.GetChange("rule")

	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oldSet := os.(*schema.Set)
	newSet := ns.(*schema.Set)

	setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
		t, ok := resource.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
		}
		// NOTE: WAF rule block doesn't have a "name" attribute.
		return t["modsec_rule_id"], nil
	})

	diffResult, err := setDiff.Diff(oldSet, newSet)
	if err != nil {
		return err
	}

	// NOTE: Fastly WAF (WAF 2020) API doesn't have a proper batch update endpoint.
	// go-fastly uses the below endpoint for UpsertBatchOperation:
	// "POST /waf/firewalls/:firewall_id/versions/:version_id/active-rules"
	// but this endpoint only updates "status" when it comes to upsert operation and "revision" field is ignored.
	// Therefore, when one of the rule attributes is changed we must delete it first and create it as a new rule.

	log.Print("[INFO] WAF rules update")
	// DELETE removed rules
	if len(diffResult.Deleted) > 0 || len(diffResult.Modified) > 0 {
		var items []interface{}
		items = append(items, diffResult.Deleted...)
		items = append(items, diffResult.Modified...)
		deleteOpts := buildBatchDeleteWAFActiveRulesInput(items, wafID, Number)
		log.Printf("[DEBUG] WAF rules delete opts: %#v", deleteOpts)
		err := executeBatchWAFActiveRulesOperations(conn, &deleteOpts)
		if err != nil {
			return err
		}
	}

	// CREATE new rules
	if len(diffResult.Added) > 0 || len(diffResult.Modified) > 0 {
		var items []interface{}
		items = append(items, diffResult.Added...)
		items = append(items, diffResult.Modified...)
		createOpts := buildBatchCreateWAFActiveRulesInput(items, wafID, Number)
		log.Printf("[DEBUG] WAF rules create opts: %#v", createOpts)
		err := executeBatchWAFActiveRulesOperations(conn, &createOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func readWAFRules(meta interface{}, d *schema.ResourceData, v int) error {

	conn := meta.(*FastlyClient).conn
	wafID := d.Get("waf_id").(string)

	log.Printf("[INFO] retrieving active rules for WAF: %s", wafID)
	resp, err := conn.ListAllWAFActiveRules(&gofastly.ListAllWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: v,
	})
	if err != nil {
		return err
	}

	rules := flattenWAFActiveRules(resp.Items)

	if err := d.Set("rule", rules); err != nil {
		log.Printf("[WARN] Error setting WAF rules for (%s): %s", d.Id(), err)
	}
	return nil
}

func buildBatchCreateWAFActiveRulesInput(items []interface{}, wafID string, wafVersionNumber int) gofastly.BatchModificationWAFActiveRulesInput {

	rules := make([]*gofastly.WAFActiveRule, len(items))
	for i, rRaw := range items {
		rf := rRaw.(map[string]interface{})

		rules[i] = &gofastly.WAFActiveRule{
			ModSecID: rf["modsec_rule_id"].(int),
			Revision: rf["revision"].(int),
			Status:   rf["status"].(string),
		}
	}

	return gofastly.BatchModificationWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Rules:            rules,
		OP:               gofastly.UpsertBatchOperation,
	}
}

func buildBatchDeleteWAFActiveRulesInput(items []interface{}, wafID string, wafVersionNumber int) gofastly.BatchModificationWAFActiveRulesInput {

	rules := make([]*gofastly.WAFActiveRule, len(items))
	for i, rRaw := range items {
		rf := rRaw.(map[string]interface{})

		rules[i] = &gofastly.WAFActiveRule{
			ModSecID: rf["modsec_rule_id"].(int),
		}
	}

	return gofastly.BatchModificationWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Rules:            rules,
		OP:               gofastly.DeleteBatchOperation,
	}
}

func executeBatchWAFActiveRulesOperations(conn *gofastly.Client, input *gofastly.BatchModificationWAFActiveRulesInput) error {

	batchSize := gofastly.WAFBatchModifyMaximumOperations
	items := input.Rules

	for i := 0; i < len(items); i += batchSize {
		j := i + batchSize
		if j > len(items) {
			j = len(items)
		}

		batch := items[i:j]

		if _, err := conn.BatchModificationWAFActiveRules(&gofastly.BatchModificationWAFActiveRulesInput{
			WAFID:            input.WAFID,
			WAFVersionNumber: input.WAFVersionNumber,
			Rules:            batch,
			OP:               input.OP,
		}); err != nil {
			return err
		}
	}
	return nil
}

func flattenWAFActiveRules(rules []*gofastly.WAFActiveRule) []map[string]interface{} {
	rl := make([]map[string]interface{}, len(rules))
	for i, r := range rules {

		ruleMapString := map[string]interface{}{
			"modsec_rule_id": r.ModSecID,
			"revision":       r.Revision,
			"status":         r.Status,
		}

		rl[i] = ruleMapString
	}
	return rl
}
