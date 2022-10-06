package fastly

import (
	"context"
	"time"

	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFastlyTLSActivation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFastlyTLSActivationRead,
		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
				Description:   "ID of the TLS Certificate used.",
			},
			"configuration_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
				Description:   "ID of the TLS Configuration used.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (GMT) when TLS was enabled.",
			},
			"domain": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
				Description:   "Domain that TLS was enabled on.",
			},
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Fastly Activation ID. Conflicts with all other filters.",
				ConflictsWith: []string{"certificate_id", "configuration_id", "domain"},
			},
		},
	}
}

func dataSourceFastlyTLSActivationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var activation *fastly.TLSActivation

	if v, ok := d.GetOk("id"); ok {
		foundActivation, err := conn.GetTLSActivation(&fastly.GetTLSActivationInput{
			ID: v.(string),
		})
		if err != nil {
			return diag.FromErr(err)
		}
		activation = foundActivation
	} else {
		filters := getTLSActivationFilters(d)

		activations, err := listTLSActivations(conn, filters...)
		if err != nil {
			return diag.FromErr(err)
		}

		if len(activations) == 0 {
			return diag.Errorf("your query returned no results. Please change your search criteria and try again")
		}

		if len(activations) > 1 {
			return diag.Errorf("your query returned more than one result. Please change to a more specific search criteria")
		}

		activation = activations[0]
	}

	err := dataSourceFastlyTLSActivationSetAttributes(activation, d)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// TLSActivationPredicate determines if an activation should be filtered.
type TLSActivationPredicate func(activation *fastly.TLSActivation) bool

func getTLSActivationFilters(d *schema.ResourceData) []TLSActivationPredicate {
	var filters []TLSActivationPredicate

	if v, ok := d.GetOk("certificate_id"); ok {
		filters = append(filters, func(c *fastly.TLSActivation) bool {
			return c.Certificate.ID == v.(string)
		})
	}
	if v, ok := d.GetOk("configuration_id"); ok {
		filters = append(filters, func(c *fastly.TLSActivation) bool {
			return c.Configuration.ID == v.(string)
		})
	}
	if v, ok := d.GetOk("domain"); ok {
		filters = append(filters, func(c *fastly.TLSActivation) bool {
			return c.Domain.ID == v.(string)
		})
	}

	return filters
}

func listTLSActivations(conn *fastly.Client, filters ...TLSActivationPredicate) ([]*fastly.TLSActivation, error) {
	var activations []*fastly.TLSActivation
	pageNumber := 1
	for {
		list, err := conn.ListTLSActivations(&fastly.ListTLSActivationsInput{
			PageNumber: pageNumber,
			PageSize:   10,
		})
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, activation := range list {
			if filterTLSActivations(activation, filters) {
				activations = append(activations, activation)
			}
		}
	}

	return activations, nil
}

func dataSourceFastlyTLSActivationSetAttributes(activation *fastly.TLSActivation, d *schema.ResourceData) error {
	d.SetId(activation.ID)

	if err := d.Set("certificate_id", activation.Certificate.ID); err != nil {
		return err
	}
	if err := d.Set("configuration_id", activation.Configuration.ID); err != nil {
		return err
	}
	if err := d.Set("domain", activation.Domain.ID); err != nil {
		return err
	}
	return d.Set("created_at", activation.CreatedAt.Format(time.RFC3339))
}

func filterTLSActivations(config *fastly.TLSActivation, filters []TLSActivationPredicate) bool {
	for _, f := range filters {
		if !f(config) {
			return false
		}
	}
	return true
}
