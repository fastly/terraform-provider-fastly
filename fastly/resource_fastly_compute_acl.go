package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/computeacls"
)

func resourceFastlyComputeACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyComputeACLCreate,
		ReadContext:   resourceFastlyComputeACLRead,
		DeleteContext: resourceFastlyComputeACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name to identify the Compute ACL. It is important to note that changing this attribute will delete and recreate the Compute ACL, and discard the current entries. You MUST first delete the associated resource_link block from your service before modifying this field.",
				ForceNew:    true,
			},
		},
	}
}

func resourceFastlyComputeACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := computeacls.CreateInput{
		Name: gofastly.ToPointer(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] CREATE: Compute ACL input: %#v", i)

	acl, err := computeacls.Create(ctx, conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(acl.ComputeACLID)

	return nil
}

func resourceFastlyComputeACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := computeacls.DescribeInput{
		ComputeACLID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] REFRESH: Compute ACL input: %#v", i)

	acl, err := computeacls.Describe(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] No Compute ACL found '%s'", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", acl.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyComputeACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	i := computeacls.DeleteInput{
		ComputeACLID: gofastly.ToPointer(d.Id()),
	}

	log.Printf("[DEBUG] DELETE: Compute ACL input: %#v", i)

	err := computeacls.Delete(gofastly.NewContextForResourceID(ctx, d.Id()), conn, &i)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
