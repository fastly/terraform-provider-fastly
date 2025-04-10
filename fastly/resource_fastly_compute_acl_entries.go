package fastly

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyComputeACLEntries() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyComputeACLEntriesCreate,
		ReadContext:   resourceFastlyComputeACLEntriesRead,
		UpdateContext: resourceFastlyComputeACLEntriesUpdate,
		DeleteContext: resourceFastlyComputeACLEntriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAclEntriesImport,
		},
		Schema: map[string]*schema.Schema{
			"acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the ACL that the entries belong to",
			},
			"entry": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "ACL Entries",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("manage_entries").(bool)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:        schema.TypeString,
							Description: "The action to take on the entry.  Valid values are `allow` or `block`",
							Required:    true,
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								if v != "ALLOW" && v != "BLOCK" {
									errs = append(errs, fmt.Errorf("%q must be either `ALLOW` or `BLOCK` case sensitive, got: %q", key, v))
								}
								return
							},
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "The ACL entry prefix in Classless Inter-Domain Routing (CIDR) notation",
							Required:    true,
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								if _, _, err := net.ParseCIDR(v); err != nil {
									errs = append(errs, fmt.Errorf("%q must be a valid CIDR notation, got: %q", key, v))
								}
								return
							},
						},
					},
				},
			},
			"force_destroy": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Allow all ACL entries to be deleted during destroy. Defaults to false.",
			},
			"manage_entries": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Whether to reapply changes if the state of the entries drifts, i.e. if entries are managed externally",
			},
		},
	}
}

func resourceFastlyComputeACLEntriesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	// Set the ID right away
	d.SetId(aclID)

	// Don't process entries if there are none to process
	if entries.Len() == 0 {
		return resourceFastlyComputeACLEntriesRead(ctx, d, meta)
	}

	var batchEntries []*computeacls.BatchComputeACLEntry = []*computeacls.BatchComputeACLEntry{}
	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)

		prefix := val["prefix"].(string)
		batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
			Operation: gofastly.ToPointer("create"),
			Prefix:    gofastly.ToPointer(val["prefix"].(string)),
			Action:    gofastly.ToPointer(val["action"].(string)),
		})

		log.Printf("[DEBUG] Creating ACL entry for ACL %s: %s", aclID, prefix)
	}
	err := computeacls.Update(conn, &computeacls.UpdateInput{
		ComputeACLID: gofastly.ToPointer(aclID),
		Entries:      batchEntries,
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating ACL entries for ACL %s: %w", aclID, err))
	}

	return resourceFastlyComputeACLEntriesRead(ctx, d, meta)
}

func resourceFastlyComputeACLEntriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)

	// We can use the computeacls package for listing entries
	input := &computeacls.ListEntriesInput{
		ComputeACLID: gofastly.ToPointer(aclID),
	}

	results, err := computeacls.ListEntries(conn, input)
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.StatusCode == 404 {
			// ACL was deleted
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error listing ACL entries for ACL %s: %w", aclID, err))
	}

	if results == nil || len(results.Entries) == 0 {
		err := d.Set("entry", []map[string]any{})
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	entries := flattenComputeACLEntries(results.Entries)
	err = d.Set("entry", entries)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlyComputeACLEntriesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)

	if d.HasChange("entry") {
		old, new := d.GetChange("entry")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		var batchEntries []*computeacls.BatchComputeACLEntry = []*computeacls.BatchComputeACLEntry{}

		// Find entries to remove
		for _, vRaw := range oldSet.Difference(newSet).List() {
			val := vRaw.(map[string]any)

			prefix := val["prefix"].(string)
			batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
				Operation: gofastly.ToPointer("delete"),
				Prefix:    gofastly.ToPointer(val["prefix"].(string)),
			})

			log.Printf("[DEBUG] Creating ACL entry for ACL %s: %s", aclID, prefix)
		}

		// Find entries to add
		for _, vRaw := range newSet.Difference(oldSet).List() {
			val := vRaw.(map[string]any)

			prefix := val["prefix"].(string)
			batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
				Operation: gofastly.ToPointer("create"),
				Prefix:    gofastly.ToPointer(val["prefix"].(string)),
				Action:    gofastly.ToPointer(val["action"].(string)),
			})

			log.Printf("[DEBUG] Creating ACL entry for ACL %s: %s", aclID, prefix)
		}

		// Find entries to update - need to delete and recreate since direct update isn't supported
		for _, vRaw := range newSet.Intersection(oldSet).List() {
			newVal := vRaw.(map[string]any)

			// Find the old entry with the same prefix
			var oldAction string
			for _, oRaw := range oldSet.List() {
				oldVal := oRaw.(map[string]any)
				if oldVal["prefix"].(string) == newVal["prefix"].(string) {
					oldAction = oldVal["action"].(string)
					break
				}
			}

			// Only update if the action has changed
			if oldAction != newVal["action"].(string) {
				// Delete the old entry first
				batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
					Operation: gofastly.ToPointer("delete"),
					Prefix:    gofastly.ToPointer(newVal["prefix"].(string)),
				})

				// Then create a new entry with the updated action
				batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
					Operation: gofastly.ToPointer("create"),
					Prefix:    gofastly.ToPointer(newVal["prefix"].(string)),
					Action:    gofastly.ToPointer(newVal["action"].(string)),
				})

				log.Printf("[DEBUG] Updating ACL entry for ACL %s: %s", aclID, newVal["prefix"].(string))
			}
		}

		err := computeacls.Update(conn, &computeacls.UpdateInput{
			ComputeACLID: gofastly.ToPointer(aclID),
			Entries:      batchEntries,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating ACL entries for ACL %s: %w", aclID, err))
		}
	}

	return resourceFastlyComputeACLEntriesRead(ctx, d, meta)
}

func resourceFastlyComputeACLEntriesDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)

	entries := d.Get("entry").(*schema.Set)
	if entries.Len() != 0 && !d.Get("force_destroy").(bool) {
		return diag.Errorf("refusing to delete ACL entries for ACL %s without force_destroy set to true", aclID)
	}

	var batchEntries []*computeacls.BatchComputeACLEntry = []*computeacls.BatchComputeACLEntry{}
	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)
		prefix := val["prefix"].(string)

		log.Printf("[DEBUG] Deleting ACL entry %s from ACL %s", prefix, aclID)

		batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
			Operation: gofastly.ToPointer("delete"),
			Prefix:    gofastly.ToPointer(prefix),
		})
	}

	// Use the regular API to delete entries
	err := computeacls.Update(conn, &computeacls.UpdateInput{
		ComputeACLID: gofastly.ToPointer(aclID),
		Entries:      batchEntries,
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.StatusCode == 404 {
			// ACL was already deleted
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting ACL entries for ACL %s: %w", aclID, err))
	}

	d.SetId("")
	return nil
}

func resourceAclEntriesImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("invalid id: %s. The ID should be in the format [acl_id]/entries", d.Id())
	}

	aclID := split[0]

	err := d.Set("acl_id", aclID)
	if err != nil {
		return nil, fmt.Errorf("error setting ACL ID (%s): %s", aclID, err)
	}

	return []*schema.ResourceData{d}, nil
}

// flattenComputeACLEntries converts API response to format suitable for Terraform state
func flattenComputeACLEntries(entries []computeacls.ComputeACLEntry) []map[string]any {
	result := make([]map[string]any, 0, len(entries))

	for _, entry := range entries {
		// Create the map with the basic info
		item := map[string]any{
			"prefix": entry.Prefix,
			"action": entry.Action,
		}

		result = append(result, item)
	}

	return result
}
