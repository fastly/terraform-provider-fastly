package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyACLCreate,
		ReadContext:   resourceFastlyACLRead,
		UpdateContext: resourceFastlyACLUpdate,
		DeleteContext: resourceFastlyACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"acl_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the ACL",
			},
			"force_destroy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Allow the ACL to be deleted, even if it contains entries. Defaults to false.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A unique name to identify the ACL.",
			},
		},
	}
}

func resourceFastlyACLCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &computeacls.CreateInput{
		Name: gofastly.ToPointer(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] CREATE: Compute ACL input: %#v", input)

	acl, err := computeacls.Create(conn, input)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the ACL ID
	if acl.ComputeACLID != "" {
		d.Set("acl_id", acl.ComputeACLID)
		d.SetId(acl.ComputeACLID) // Using just the ACL ID as the resource ID
	}

	return nil
}

func resourceFastlyACLRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	aclID := d.Id()

	// Get the ACL by ID
	acl, err := computeacls.Describe(conn, &computeacls.DescribeInput{
		ComputeACLID: gofastly.ToPointer(aclID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.StatusCode == 404 {
			// ACL was deleted
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("acl_id", acl.ComputeACLID)
	d.Set("name", acl.Name)

	return nil
}

func resourceFastlyACLUpdate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// The Update operation doesn't support updating name directly in the computeacls API
	// If we need to support this, we would need to do a Delete and Create operation
	// For now, we'll return an error if name changes
	if d.HasChange("name") {
		return diag.Errorf("changing the name of an ACL is not currently supported")
	}

	return nil
}

func resourceFastlyACLDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Id()

	// Check if force_destroy is set
	if !d.Get("force_destroy").(bool) {
		// Check if the ACL has any entries
		entries, err := computeacls.ListEntries(conn, &computeacls.ListEntriesInput{
			ComputeACLID: gofastly.ToPointer(aclID),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error checking ACL entries: %w", err))
		}

		if entries != nil && len(entries.Entries) > 0 {
			return diag.FromErr(fmt.Errorf("cannot delete ACL (%s), list is not empty. Either delete the entries first, or set force_destroy to true and apply it before making this change", aclID))
		}
	}

	// Delete the ACL
	err := computeacls.Delete(conn, &computeacls.DeleteInput{
		ComputeACLID: gofastly.ToPointer(aclID),
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.StatusCode == 404 {
			// ACL was already deleted
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
