package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"

	wsr "github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/redactions"
)

func dataSourceFastlyNGWAFRedactions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyNGWAFRedactionsRead,
		Schema: map[string]*schema.Schema{
			"redactions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all redactions for a workspace.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the field that is being redacted.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the redaction.",
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of field being redacted. One of `request_parameter`, `request_header`, or `response_header`.",
							Computed:    true,
						},
					},
				},
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Description: "The ID of the workspace.",
				Required:    true,
			},
		},
	}
}

func dataSourceFastlyNGWAFRedactionsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	workspaceID := d.Get("workspace_id").(string)

	log.Printf("[DEBUG] Reading NGWAF redactions from workspace %s", workspaceID)

	remoteState, err := wsr.List(ctx, conn, &wsr.ListInput{
		WorkspaceID: &workspaceID,
	})
	if err != nil {
		return diag.Errorf("error fetching redactions: %s", err)
	}

	parsed, _ := json.Marshal(remoteState)
	hash := strconv.Itoa(hashcode.String(string(parsed)))
	d.SetId(hash)

	// Convert []redactions.Redaction to []*redactions.Redaction
	var redactionPtrs []*wsr.Redaction
	for i := range remoteState.Data {
		redactionPtrs = append(redactionPtrs, &remoteState.Data[i])
	}

	if err := d.Set("redactions", flattenNGWAFRedactions(redactionPtrs)); err != nil {
		return diag.Errorf("error setting redactions: %s", err)
	}

	return nil
}

func flattenNGWAFRedactions(remoteState []*wsr.Redaction) []map[string]any {
	result := make([]map[string]any, len(remoteState))

	for i, r := range remoteState {
		result[i] = map[string]any{
			"id":    r.RedactionID,
			"field": r.Field,
			"type":  r.Type,
		}
	}

	return result
}
