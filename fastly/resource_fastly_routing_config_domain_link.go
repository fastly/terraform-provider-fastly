package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/domains"
)

func resourceFastlyRoutingConfigDomainLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyRoutingConfigDomainLinkUpdate,
		ReadContext:   resourceFastlyRoutingConfigDomainLinkRead,
		UpdateContext: resourceFastlyRoutingConfigDomainLinkUpdate,
		DeleteContext: resourceFastlyRoutingConfigDomainLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyRoutingConfigDomainLinkImport,
		},

		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Domain Identifier of the versionless domain being linked (UUID).",
			},
			"routing_config_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The routing config identifier to link to the domain.",
			},
		},
	}
}

func resourceFastlyRoutingConfigDomainLinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Routing Config Domain Link for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	input := &domains.GetInput{
		DomainID: new(d.Get("domain_id").(string)),
	}

	data, err := domains.Get(gofastly.NewContextForResourceID(ctx, d.Get("domain_id").(string)), conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] Domain (%s) not found, removing routing config domain link from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if data.RoutingConfigurationID == nil {
		log.Printf("[WARN] Domain (%s) has no routing config linked, removing routing config domain link from state", d.Id())
		d.SetId("")
		return nil
	}
	if err := d.Set("domain_id", data.DomainID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("routing_config_id", data.RoutingConfigurationID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(data.DomainID)

	return nil
}

func resourceFastlyRoutingConfigDomainLinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &domains.UpdateInput{
		DomainID:               new(d.Get("domain_id").(string)),
		RoutingConfigurationID: gofastly.NewNullable(d.Get("routing_config_id").(string)),
	}
	_, err := domains.Update(gofastly.NewContextForResourceID(ctx, d.Get("domain_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyRoutingConfigDomainLinkRead(ctx, d, meta)
}

func resourceFastlyRoutingConfigDomainLinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &domains.UpdateInput{
		DomainID:               new(d.Id()),
		RoutingConfigurationID: gofastly.NullValue[string](),
	}
	_, err := domains.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); !ok || !e.IsNotFound() {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}

func resourceFastlyRoutingConfigDomainLinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*APIClient).conn
	domainID := d.Id()

	input := &domains.GetInput{
		DomainID: new(domainID),
	}
	data, err := domains.Get(ctx, conn, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	if data.RoutingConfigurationID == nil {
		return nil, fmt.Errorf("domain %s has no routing_config_id set and cannot be imported", domainID)
	}

	if err := d.Set("domain_id", domainID); err != nil {
		return nil, err
	}
	if err := d.Set("routing_config_id", data.RoutingConfigurationID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
