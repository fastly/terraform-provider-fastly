package fastly

import (
	"context"
	"fmt"
	"log"
	"strings"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyACLEntries() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlyACLEntriesCreate,
		ReadContext:   resourceFastlyACLEntriesRead,
		UpdateContext: resourceFastlyACLEntriesUpdate,
		DeleteContext: resourceFastlyACLEntriesDelete,
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
			"entries": {
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
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "The ACL entry prefix in Classless Inter-Domain Routing (CIDR) notation",
							Required:    true,
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

func resourceFastlyACLEntriesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	// Set the ID right away
	d.SetId(aclID)

	// Don't process entries if there are none to process
	if entries.Len() == 0 {
		return resourceFastlyACLEntriesRead(ctx, d, meta)
	}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)

		// We need to create an entry with the ACL resource
		createEntry := &gofastly.CreateACLEntryInput{
			ServiceID: "", // Not used for compute ACLs
			ACLID:     aclID,
			IP:        gofastly.ToPointer(val["ip"].(string)),
			Negated:   gofastly.ToPointer(gofastly.Compatibool(val["negated"].(bool))),
		}

		if comment, ok := val["comment"].(string); ok && comment != "" {
			createEntry.Comment = gofastly.ToPointer(comment)
		}

		if subnet, ok := val["subnet"].(string); ok && subnet != "" {
			subnetInt := computeConvertSubnetToInt(subnet)
			createEntry.Subnet = gofastly.ToPointer(subnetInt)
		}

		log.Printf("[DEBUG] Creating ACL entry for ACL %s: %#v", aclID, createEntry)
		entry, err := conn.CreateACLEntry(createEntry)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating ACL entry for ACL %s: %w", aclID, err))
		}
		log.Printf("[DEBUG] Created ACL entry %s for ACL %s", *entry.EntryID, aclID)
	}

	return resourceFastlyACLEntriesRead(ctx, d, meta)
}

func resourceFastlyACLEntriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func resourceFastlyACLEntriesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)

	if d.HasChange("entry") {
		old, new := d.GetChange("entry")
		oldSet := old.(*schema.Set)
		newSet := new.(*schema.Set)

		var batchEntries []*computeacls.BatchComputeACLEntry

		// Find entries to remove
		for _, vRaw := range oldSet.Difference(newSet).List() {
			val := vRaw.(map[string]any)

			prefix := val["prefix"].(string)
			batchEntries = append(batchEntries, &computeacls.BatchComputeACLEntry{
				Operation: gofastly.ToPointer("delete"),
				Prefix:    gofastly.ToPointer(val["prefix"].(string)),
				Action:    gofastly.ToPointer(val["action"].(string)),
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
					Action:    gofastly.ToPointer(oldAction),
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

		computeacls.Update(conn, &computeacls.UpdateInput{
			ComputeACLID: gofastly.ToPointer(aclID),
			Entries:      batchEntries,
		})
	}

	return resourceFastlyACLEntriesRead(ctx, d, meta)
}

func resourceFastlyACLEntriesDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn
	aclID := d.Get("acl_id").(string)

	if !d.Get("force_destroy").(bool) {
		return diag.Errorf("refusing to delete ACL entries for ACL %s without force_destroy set to true", aclID)
	}

	entries := d.Get("entry").(*schema.Set)
	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)
		entryID := val["id"].(string)

		log.Printf("[DEBUG] Deleting ACL entry %s from ACL %s", entryID, aclID)

		// Use the regular API to delete entries
		err := conn.DeleteACLEntry(&gofastly.DeleteACLEntryInput{
			ServiceID: "", // Not used for compute ACLs
			ACLID:     aclID,
			EntryID:   entryID,
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error deleting ACL entry %s from ACL %s: %w", entryID, aclID, err))
		}
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

func computeConvertSubnetToInt(s string) int {
	subnet := 0
	if s != "" {
		fmt.Sscanf(s, "%d", &subnet)
	}
	return subnet
}
