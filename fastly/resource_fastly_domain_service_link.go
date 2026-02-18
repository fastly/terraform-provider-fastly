package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/domainmanagement/v1/domains"
)

func resourceFastlyDomainServiceLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyDomainServiceLinkUpdate,
		ReadContext:   resourceFastlyDomainServiceLinkRead,
		UpdateContext: resourceFastlyDomainServiceLinkUpdate,
		DeleteContext: resourceFastlyDomainServiceLinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyDomainServiceLinkImport,
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

func resourceFastlyDomainServiceLinkImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*APIClient).conn
	domainID := d.Id()

	// Fetch the domain to get service_id
	input := &domains.GetInput{
		DomainID: gofastly.ToPointer(domainID),
	}
	data, err := domains.Get(ctx, conn, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}
	if data.ServiceID == nil {
		return nil, fmt.Errorf("domain %s has no service_id set and cannot be imported", domainID)
	}

	// Set both domain_id and service_id
	if err := d.Set("domain_id", domainID); err != nil {
		return nil, err
	}
	if err := d.Set("service_id", data.ServiceID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// Deprecated resources.
func resourceFastlyDomainServiceLinkV1() *schema.Resource {
	resource := resourceFastlyDomainServiceLink()
	resource.DeprecationMessage = "This resource is deprecated. Please use 'fastly_domain_service_link' instead."
	return resource
}
