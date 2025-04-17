package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyVCLSnippets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyVCLSnippetsRead,
		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Alphanumeric string identifying the service.",
			},
			"service_version": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Integer identifying a service version.",
			},
			"vcl_snippets": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of all VCL snippets for the version of the service.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The VCL code that specifies exactly what the snippet does.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying a VCL Snippet.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name for the snippet.	",
						},
						"priority": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Priority determines execution order. Lower numbers execute first.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The location in generated VCL where the snippet should be placed.",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyVCLSnippetsRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading VCL snippets")

	remoteState, err := conn.ListSnippets(&gofastly.ListSnippetsInput{
		ServiceID:      d.Get("service_id").(string),
		ServiceVersion: d.Get("service_version").(int),
	})
	if err != nil {
		return diag.Errorf("error fetching VCL snippets: %s", err)
	}

	hashBase, _ := json.Marshal(remoteState)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("vcl_snippets", flattenDataSourceVCLSnippets(remoteState)); err != nil {
		return diag.Errorf("error setting vcl_snippets: %s", err)
	}

	return nil
}

// flattenDataSourceVCLSnippets models data into format suitable for saving to
// Terraform state.
func flattenDataSourceVCLSnippets(remoteState []*gofastly.Snippet) []map[string]any {
	result := make([]map[string]any, len(remoteState))
	if len(remoteState) == 0 {
		return result
	}

	for i, resource := range remoteState {
		result[i] = map[string]any{}

		if resource.Content != nil {
			result[i]["content"] = *resource.Content
		}
		if resource.SnippetID != nil {
			result[i]["id"] = *resource.SnippetID
		}
		if resource.Name != nil {
			result[i]["name"] = *resource.Name
		}
		if resource.Priority != nil {
			p, _ := strconv.Atoi(*resource.Priority)
			result[i]["priority"] = p
		}
		if resource.Type != nil {
			result[i]["type"] = *resource.Type
		}
	}

	return result
}
