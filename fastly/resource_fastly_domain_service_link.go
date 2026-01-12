package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/domainmanagement/v1/domains"
)

func resourceFastlyDomainServiceLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyDomainServiceLinkUpdate,
		ReadContext:   resourceFastlyDomainServiceLinkRead,
		UpdateContext: resourceFastlyDomainServiceLinkUpdate,
		DeleteContext: resourceFastlyDomainServiceLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Domain Identifier of the versionless domain being linked (UUID).",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The service_id associated with your domain",
			},
		},
	}
}

func resourceFastlyDomainServiceLinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Domain Service Link Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	input := &domains.GetInput{
		DomainID: gofastly.ToPointer(d.Get("domain_id").(string)),
	}

	data, err := domains.Get(gofastly.NewContextForResourceID(ctx, d.Get("domain_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain_id", data.DomainID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_id", data.ServiceID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(data.DomainID)

	return nil
}

func resourceFastlyDomainServiceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &domains.UpdateInput{
		DomainID:  gofastly.ToPointer(d.Get("domain_id").(string)),
		ServiceID: gofastly.ToPointer(d.Get("service_id").(string)),
	}
	_, err := domains.Update(gofastly.NewContextForResourceID(ctx, d.Get("domain_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyDomainServiceLinkRead(ctx, d, meta)
}

func resourceFastlyDomainServiceLinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &domains.UpdateInput{
		DomainID:  gofastly.ToPointer(d.Id()),
		ServiceID: nil,
	}
	_, err := domains.Update(gofastly.NewContextForResourceID(ctx, d.Id()), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyDomainServiceLinkRead(ctx, d, meta)
}
