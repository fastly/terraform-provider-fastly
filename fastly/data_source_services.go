package fastly

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/fastly/terraform-provider-fastly/fastly/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyServicesRead,
		Schema: map[string]*schema.Schema{
			"details": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A detailed list of Fastly services in your account. This is limited to the services the API token can read.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A freeform descriptive note.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date and time in ISO 8601 format.",
						},
						"customer_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the customer.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Alphanumeric string identifying the service.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the service.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of this service. One of `vcl`, `wasm`.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date and time in ISO 8601 format.",
						},
						"version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The currently activated version.",
						},
					},
				},
			},
			"ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A list of service IDs in your account. This is limited to the services the API token can read.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceFastlyServicesRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] Reading services")

	services, err := conn.ListServices(&gofastly.ListServicesInput{})
	if err != nil {
		return diag.Errorf("error fetching services: %s", err)
	}

	hashBase, _ := json.Marshal(services)
	hashString := strconv.Itoa(hashcode.String(string(hashBase)))
	d.SetId(hashString)

	if err := d.Set("details", flattenServiceDetails(services)); err != nil {
		return diag.Errorf("error setting services: %s", err)
	}

	if err := d.Set("ids", flattenServiceIDs(services)); err != nil {
		return diag.Errorf("error setting service IDs: %s", err)
	}

	return nil
}

func flattenServiceIDs(services []*gofastly.Service) []string {
	result := make([]string, len(services))
	for i, s := range services {
		result[i] = s.ID
	}
	return result
}

func flattenServiceDetails(services []*gofastly.Service) []map[string]any {
	result := make([]map[string]any, len(services))
	if len(services) == 0 {
		return result
	}

	for i, s := range services {
		result[i] = map[string]any{
			"id":          s.ID,
			"name":        s.Name,
			"type":        s.Type,
			"comment":     s.Comment,
			"customer_id": s.CustomerID,
			"created_at":  s.CreatedAt.String(),
			"updated_at":  s.UpdatedAt.String(),
			"version":     s.ActiveVersion,
		}
	}

	return result
}
