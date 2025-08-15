package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
	"github.com/fastly/go-fastly/v11/fastly/domainmanagement/v1/domains"
)

func resourceFastlyDomainV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyDomainV1Create,
		ReadContext:   resourceFastlyDomainV1Read,
		UpdateContext: resourceFastlyDomainV1Update,
		DeleteContext: resourceFastlyDomainV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for your domain.",
			},
			"domain_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Domain Identifier (UUID).",
			},
			"fqdn": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The fully-qualified domain name for your domain (e.g. `www.example.com`, no trailing dot). Can be created, but not updated.",
			},
			"service_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The service_id associated with your domain or null if there is no association.",
			},
		},
	}
}

func resourceFastlyDomainV1Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var input domains.CreateInput
	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("fqdn"); ok {
		input.FQDN = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("service_id"); ok {
		input.ServiceID = gofastly.ToPointer(v.(string))
	}

	data, err := domains.Create(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), conn, &input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(data.DomainID)

	if err := d.Set("domain_id", data.DomainID); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceFastlyDomainV1Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Domain V1 Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	input := &domains.GetInput{
		DomainID: gofastly.ToPointer(d.Id()),
	}

	data, err := domains.Get(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("description", data.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain_id", data.DomainID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fqdn", data.FQDN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_id", data.ServiceID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyDomainV1Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &domains.UpdateInput{
		DomainID: gofastly.ToPointer(d.Id()),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("service_id"); ok {
		input.ServiceID = gofastly.ToPointer(v.(string))
	}

	_, err := domains.Update(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyDomainV1Read(ctx, d, meta)
}

func resourceFastlyDomainV1Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	input := &domains.DeleteInput{
		DomainID: gofastly.ToPointer(d.Id()),
	}
	err := domains.Delete(gofastly.NewContextForResourceID(ctx, d.Get("service_id").(string)), conn, input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
