package fastly

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
)

func resourceFastlyAPISecurityOperationTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAPISecurityOperationTagCreate,
		ReadContext:   resourceFastlyAPISecurityOperationTagRead,
		UpdateContext: resourceFastlyAPISecurityOperationTagUpdate,
		DeleteContext: resourceFastlyAPISecurityOperationTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyAPISecurityOperationTagImport,
		},

		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Created timestamp (when present).",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the operation tag.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the operation tag.",
			},
			"operation_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of operations associated with this tag (when present).",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service ID the tag belongs to. To import, use: <service_id>/<tag_id>.",
			},
			"tag_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tag ID.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Updated timestamp (when present).",
			},
		},
	}
}

func resourceFastlyAPISecurityOperationTagCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	serviceID := d.Get("service_id").(string)

	in := &operations.CreateTagInput{
		ServiceID: gofastly.ToPointer(serviceID),
		Name:      gofastly.ToPointer(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		in.Description = gofastly.ToPointer(v.(string))
	}

	log.Printf("[DEBUG] Creating API Security operation tag: %#v", in)
	tag, err := operations.CreateTag(gofastly.NewContextForResourceID(ctx, serviceID), conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("tag_id", tag.ID)
	d.SetId(fmt.Sprintf("%s/%s", serviceID, tag.ID))

	return resourceFastlyAPISecurityOperationTagRead(ctx, d, meta)
}

func resourceFastlyAPISecurityOperationTagRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing API Security Operation Tag for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	serviceID, tagID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	tag, err := operations.DescribeTag(ctx, conn, &operations.DescribeTagInput{
		ServiceID: gofastly.ToPointer(serviceID),
		TagID:     gofastly.ToPointer(tagID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("created_at", tag.CreatedAt)
	_ = d.Set("description", tag.Description)
	_ = d.Set("name", tag.Name)
	_ = d.Set("operation_count", tag.Count)
	_ = d.Set("service_id", serviceID)
	_ = d.Set("tag_id", tag.ID)
	_ = d.Set("updated_at", tag.UpdatedAt)

	d.SetId(fmt.Sprintf("%s/%s", serviceID, tag.ID))
	return nil
}

func resourceFastlyAPISecurityOperationTagUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID, tagID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// API may require "name" on PATCH even when only changing description.
	in := &operations.UpdateTagInput{
		ServiceID: gofastly.ToPointer(serviceID),
		TagID:     gofastly.ToPointer(tagID),
		Name:      gofastly.ToPointer(d.Get("name").(string)),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			in.Description = gofastly.ToPointer(v.(string))
		} else {
			in.Description = gofastly.ToPointer("")
		}
	}

	log.Printf("[DEBUG] Updating API Security operation tag: %#v", in)
	_, err = operations.UpdateTag(gofastly.NewContextForResourceID(ctx, serviceID), conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAPISecurityOperationTagRead(ctx, d, meta)
}

func resourceFastlyAPISecurityOperationTagDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID, tagID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting API Security operation tag (%s/%s)", serviceID, tagID)
	err = operations.DeleteTag(gofastly.NewContextForResourceID(ctx, serviceID), conn, &operations.DeleteTagInput{
		ServiceID: gofastly.ToPointer(serviceID),
		TagID:     gofastly.ToPointer(tagID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyAPISecurityOperationTagImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	serviceID, tagID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return nil, err
	}

	if err := d.Set("service_id", serviceID); err != nil {
		return nil, fmt.Errorf("error setting service_id (%s): %w", serviceID, err)
	}
	if err := d.Set("tag_id", tagID); err != nil {
		return nil, fmt.Errorf("error setting tag_id (%s): %w", tagID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, tagID))
	return []*schema.ResourceData{d}, nil
}
