package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v12/fastly/domainmanagement/v1/domains"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyDomains() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyDomainsRead,
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A domain represents the domain name through which visitors will retrieve content. There can be multiple domains for a service.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fqdn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The fully-qualified domain name for your domain.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Domain Identifier (UUID).",
						},
						"service_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The 'service_id' associated with your domain or 'null' if there is no association.",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "The total number of domains returned.",
			},
		},
	}
}

func dataSourceFastlyDomainsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading all domains")

	var allDomains []domains.Data
	var cursor *string
	// We are setting to the limit to it's max possible
	// value (100) to return as many results as possible.
	limit := 100

	for {
		remoteState, err := domains.List(ctx, conn, &domains.ListInput{
			Limit:  &limit,
			Cursor: cursor,
		})
		if err != nil {
			return diag.Errorf("error fetching domains: %s", err)
		}

		// Capture all domains until the cursor is empty
		// indicating there are no more results to fetch
		allDomains = append(allDomains, remoteState.Data...)

		// Check if there is another cursor (page)
		if remoteState.Meta.NextCursor == "" {
			break
		}
		cursor = &remoteState.Meta.NextCursor
	}
	// Create a collection with all results
	allResults := &domains.Collection{
		Data: allDomains,
		Meta: domains.Meta{
			// Capture total number of domains
			Total: len(allDomains),
		},
	}

	hashBase, _ := json.Marshal(allResults)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("domains", flattenDomainsVersionles(allResults)); err != nil {
		return diag.Errorf("error setting domains: %s", err)
	}

	// Optional param to set if a user wants a count associated with the data source
	if err := d.Set("total", allResults.Meta.Total); err != nil {
		return diag.Errorf("error setting total: %s", err)
	}

	return nil
}

// flattenDomainsVersionles models data into format suitable for saving to Terraform state.
func flattenDomainsVersionles(remoteState *domains.Collection) []map[string]any {
	if remoteState == nil || len(remoteState.Data) == 0 {
		return []map[string]any{}
	}

	var result []map[string]any
	for _, domain := range remoteState.Data {
		// Only include domains where FQDN exists and is not empty
		if domain.FQDN != "" {
			data := map[string]any{
				"id":   domain.DomainID,
				"fqdn": domain.FQDN,
			}
			// Add service_id only if there is a service associated
			if domain.ServiceID != nil {
				data["service_id"] = *domain.ServiceID
			}
			result = append(result, data)
		}
	}

	return result
}
