package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
)

func dataSourceFastlyStagingIPs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyStagingIPsRead,
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of domains with their staging IP addresses.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The domain name.",
						},
						"staging_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The staging IP address for the domain.",
						},
					},
				},
			},
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
		},
	}
}

func dataSourceFastlyStagingIPsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading staging IPs")

	remoteState, err := conn.ListDomains(ctx, &gofastly.ListDomainsInput{
		ServiceID:      d.Get("service_id").(string),
		ServiceVersion: d.Get("service_version").(int),
		IncludeStagingIPs: true,
	})
	if err != nil {
		return diag.Errorf("error fetching staging IPs: %s", err)
	}

	hashBase, _ := json.Marshal(remoteState)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("domains", flattenDataSourceStagingIPs(remoteState)); err != nil {
		return diag.Errorf("error setting domains: %s", err)
	}

	return nil
}

// flattenDataSourceStagingIPs models data into format suitable for saving to
// Terraform state.
func flattenDataSourceStagingIPs(remoteState []*gofastly.Domain) []map[string]any {
	result := make([]map[string]any, len(remoteState))
	if len(remoteState) == 0 {
		return result
	}

	for i, resource := range remoteState {
		result[i] = map[string]any{}

		if resource.Name != nil {
			result[i]["name"] = *resource.Name
		}
		if resource.StagingIP != nil {
			result[i]["staging_ip"] = *resource.StagingIP
		}
	}

	return result
}
