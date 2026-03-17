package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"
)

func resourceFastlyAPISecurityOperation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyAPISecurityOperationCreate,
		ReadContext:   resourceFastlyAPISecurityOperationRead,
		UpdateContext: resourceFastlyAPISecurityOperationUpdate,
		DeleteContext: resourceFastlyAPISecurityOperationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyAPISecurityOperationImport,
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
				Description: "A description of the operation.",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Domain for the operation (exact match). Can be created, but not updated.",
			},
			"last_seen_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last seen timestamp (when present).",
			},
			"method": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "HTTP method for the operation (e.g. GET, POST). Can be created, but not updated.",
			},
			"operation_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The operation ID.",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Path for the operation (exact match). Can be created, but not updated.",
			},
			"rps": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Observed requests per second (when present).",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service ID the operation belongs to. To import, use: <service_id>/<operation_id>.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Discovery status (when present).",
			},
			"tag_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Associated operation tag IDs.",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "A tag ID.",
				},
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Updated timestamp (when present).",
			},
		},
	}
}

func resourceFastlyAPISecurityOperationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	serviceID := d.Get("service_id").(string)

	in := &operations.CreateInput{
		ServiceID: gofastly.ToPointer(serviceID),
		Method:    gofastly.ToPointer(strings.ToUpper(d.Get("method").(string))),
		Domain:    gofastly.ToPointer(d.Get("domain").(string)),
		Path:      gofastly.ToPointer(d.Get("path").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = gofastly.ToPointer(v.(string))
	}
	if v, ok := d.GetOk("tag_ids"); ok {
		in.TagIDs = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating API Security operation: %#v", in)
	op, err := operations.Create(gofastly.NewContextForResourceID(ctx, serviceID), conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("operation_id", op.ID)
	d.SetId(fmt.Sprintf("%s/%s", serviceID, op.ID))

	return resourceFastlyAPISecurityOperationRead(ctx, d, meta)
}

func resourceFastlyAPISecurityOperationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing API Security Operation for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	serviceID, opID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	op, err := operations.Describe(ctx, conn, &operations.DescribeInput{
		ServiceID:   gofastly.ToPointer(serviceID),
		OperationID: gofastly.ToPointer(opID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("created_at", op.CreatedAt)
	_ = d.Set("description", op.Description)
	_ = d.Set("domain", op.Domain)
	_ = d.Set("last_seen_at", op.LastSeenAt)
	_ = d.Set("method", op.Method)
	_ = d.Set("operation_id", op.ID)
	_ = d.Set("path", op.Path)
	_ = d.Set("rps", op.RPS)
	_ = d.Set("service_id", serviceID)
	_ = d.Set("status", op.Status)
	_ = d.Set("updated_at", op.UpdatedAt)

	if err := d.Set("tag_ids", flattenStringSliceToSet(op.TagIDs)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, op.ID))
	return nil
}

func resourceFastlyAPISecurityOperationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID, opID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	in := &operations.UpdateInput{
		ServiceID:   gofastly.ToPointer(serviceID),
		OperationID: gofastly.ToPointer(opID),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			in.Description = gofastly.ToPointer(v.(string))
		} else {
			in.Description = gofastly.ToPointer("")
		}
	}

	if d.HasChange("tag_ids") {
		if v, ok := d.GetOk("tag_ids"); ok {
			in.TagIDs = expandStringSet(v.(*schema.Set))
		} else {
			in.TagIDs = []string{}
		}
	}

	log.Printf("[DEBUG] Updating API Security operation: %#v", in)
	_, err = operations.Update(gofastly.NewContextForResourceID(ctx, serviceID), conn, in)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceFastlyAPISecurityOperationRead(ctx, d, meta)
}

func resourceFastlyAPISecurityOperationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID, opID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting API Security operation (%s/%s)", serviceID, opID)
	err = operations.Delete(gofastly.NewContextForResourceID(ctx, serviceID), conn, &operations.DeleteInput{
		ServiceID:   gofastly.ToPointer(serviceID),
		OperationID: gofastly.ToPointer(opID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

func resourceFastlyAPISecurityOperationImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	serviceID, opID, err := parseTwoPartImportID(d.Id())
	if err != nil {
		return nil, err
	}

	if err := d.Set("operation_id", opID); err != nil {
		return nil, fmt.Errorf("error setting operation_id (%s): %w", opID, err)
	}
	if err := d.Set("service_id", serviceID); err != nil {
		return nil, fmt.Errorf("error setting service_id (%s): %w", serviceID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, opID))
	return []*schema.ResourceData{d}, nil
}
