package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/go-fastly/v10/fastly/computeacls"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
)

func resourceFastlyComputeACLEntries() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyComputeACLEntriesCreate,
		ReadContext:   resourceFastlyComputeACLEntriesRead,
		UpdateContext: resourceFastlyComputeACLEntriesUpdate,
		DeleteContext: resourceFastlyComputeACLEntriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFastlyComputeACLEntriesImport,
		},
		Schema: map[string]*schema.Schema{
			"compute_acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Manages entries for a Fastly Compute Access Control List (ACL). To import, use the format <compute_acl_id>/entries.",
			},
			"entries": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "A map representing the entries in the Compute ACL, where the keys are the prefixes and the values are the actions (ALLOW or BLOCK).",
				Elem:        schema.TypeString,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					return !d.HasChange("compute_acl_id") && !d.Get("manage_entries").(bool)
				},
			},
			"manage_entries": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Manage the ACL entries in Terraform (default: false). If true, Terraform will ensure that the ACL's entries match the entries in the Terraform configuration.",
			},
		},
	}
}

func resourceFastlyComputeACLEntriesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] CREATE: Compute ACL Entries")
	return resourceFastlyComputeACLEntriesUpdate(ctx, d, meta)
}

func resourceFastlyComputeACLEntriesRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] REFRESH: Compute ACL Entries")

	id := d.Get("compute_acl_id").(string)
	remoteState, err := computeacls.ListEntries(conn, &computeacls.ListEntriesInput{
		ComputeACLID: &id,
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] No Compute ACL found '%s'", id)
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	entries := flattenComputeACLEntries(remoteState.Entries)
	if err := d.Set("entries", entries); err != nil {
		return diag.FromErr(fmt.Errorf("error setting 'entries': %w", err))
	}

	return nil
}

func resourceFastlyComputeACLEntriesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	log.Printf("[DEBUG] UPDATE: Compute ACL Entries")

	id := d.Get("compute_acl_id").(string)

	oldRaw, newRaw := d.GetChange("entries")
	oldEntries := oldRaw.(map[string]any)
	newEntries := newRaw.(map[string]any)

	var batch []*computeacls.BatchComputeACLEntry
	manage := d.Get("manage_entries").(bool)

	if manage {
		for prefix := range oldEntries {
			if _, ok := newEntries[prefix]; !ok {
				prefix := prefix // avoid reference issue
				batch = append(batch, &computeacls.BatchComputeACLEntry{
					Prefix:    &prefix,
					Operation: gofastly.ToPointer("delete"),
				})
			}
		}
	}

	for prefix, val := range newEntries {
		action := val.(string)
		op := "create"
		if _, ok := oldEntries[prefix]; ok {
			op = "update"
		}
		prefix := prefix // avoid reference issue
		batch = append(batch, &computeacls.BatchComputeACLEntry{
			Prefix:    &prefix,
			Action:    &action,
			Operation: &op,
		})
	}

	log.Printf("[DEBUG] Batch updating Compute ACL entries: %+v", batch)
	if err := batchUpdateComputeACLEntries(conn, id, batch); err != nil {
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		id := d.Get("compute_acl_id").(string)
		d.SetId(fmt.Sprintf("%s/entries", id))
	}

	return resourceFastlyComputeACLEntriesRead(ctx, d, meta)
}

func resourceFastlyComputeACLEntriesDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	id := d.Get("compute_acl_id").(string)
	entries := d.Get("entries").(map[string]any)

	log.Printf("[DEBUG] DELETE: Compute ACL Entries")

	var batch []*computeacls.BatchComputeACLEntry
	for prefix := range entries {
		prefix := prefix // avoid reference issue
		batch = append(batch, &computeacls.BatchComputeACLEntry{
			Prefix:    &prefix,
			Operation: gofastly.ToPointer("delete"),
		})
	}

	if err := batchUpdateComputeACLEntries(conn, id, batch); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

// flattenComputeACLEntries models data into format suitable for saving to
// Terraform state.
func flattenComputeACLEntries(entries []computeacls.ComputeACLEntry) map[string]string {
	result := make(map[string]string)
	for _, entry := range entries {
		result[entry.Prefix] = entry.Action
	}
	return result
}

// batchUpdateComputeACLEntries sends the given batch of operations to the Fastly API.
func batchUpdateComputeACLEntries(conn *gofastly.Client, computeACLID string, entries []*computeacls.BatchComputeACLEntry) error {
	// No known documented limit yet, so no batching logic — unlike config store.
	if len(entries) == 0 {
		return nil
	}

	return computeacls.Update(conn, &computeacls.UpdateInput{
		ComputeACLID: &computeACLID,
		Entries:      entries,
	})
}

func resourceFastlyComputeACLEntriesImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	computeACLID, suffix, ok := strings.Cut(d.Id(), "/")
	if !ok || suffix != "entries" {
		return nil, fmt.Errorf("invalid ID format: %s. Expected format: <compute_acl_id>/entries", d.Id())
	}

	if err := d.Set("compute_acl_id", computeACLID); err != nil {
		return nil, fmt.Errorf("error setting compute_acl_id (%s): %w", computeACLID, err)
	}

	// Normalize the ID in case the original had redundant slashes, etc.
	d.SetId(fmt.Sprintf("%s/entries", computeACLID))

	return []*schema.ResourceData{d}, nil
}
