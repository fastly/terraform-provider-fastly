package fastly

import (
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var activeRule = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Web Application Firewall rule's status. Allowed values are (log, block and score).",
				ValidateFunc: validateRuleStatusType(),
			},
			"modsec_rule_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Web Application Firewall rule's modsec ID.",
			},
			"revision": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The Web Application Firewall rule's revision.",
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

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)

	add := nss.Difference(oss).List()
	remove := deleteByModSecID(oss.Difference(nss), add).List()

	log.Print("[INFO] WAF rules update")
	if len(remove) > 0 {
		deleteOpts := buildBatchDeleteWAFActiveRulesInput(remove, wafID, Number)
		log.Printf("[DEBUG] WAF rules delete opts: %#v", deleteOpts)
		err := executeBatchWAFActiveRulesOperations(conn, &deleteOpts)
		if err != nil {
			return err
		}
	}

	if len(add) > 0 {
		createOpts := buildBatchCreateWAFActiveRulesInput(add, wafID, Number)
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

// deleteByModSecID returns a copy of the argument "remove" with all common (with the same modsec_rule_id) elements with argument "add" removed.
func deleteByModSecID(remove *schema.Set, add []interface{}) *schema.Set {

	modSecIDs := make(map[int]interface{}, remove.Len())
	result := schema.CopySet(remove)

	for _, rv := range remove.List() {
		r := rv.(map[string]interface{})
		modSecIDs[r["modsec_rule_id"].(int)] = r
	}

	if len(modSecIDs) > 0 {
		for _, av := range add {
			a := av.(map[string]interface{})
			if v, ok := modSecIDs[a["modsec_rule_id"].(int)]; ok {
				result.Remove(v)
			}
		}
	}
	return result
}
