package fastly

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func resourceServiceACLEntries() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceACLEntriesCreate,
		ReadContext:   resourceServiceACLEntriesRead,
		UpdateContext: resourceServiceACLEntriesUpdate,
		DeleteContext: resourceServiceACLEntriesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceACLEntriesImport,
		},
		Schema: map[string]*schema.Schema{
			"acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the ACL that the items belong to",
			},
			"entry": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "ACL Entries",
				MaxItems:    gofastly.MaximumACLSize,
				DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
					return !d.HasChange("acl_id") && !d.Get("manage_entries").(bool)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A personal freeform descriptive note",
						},
						"id": {
							Type:        schema.TypeString,
							Description: "The unique ID of the entry",
							Computed:    true,
						},
						"ip": {
							Type:        schema.TypeString,
							Description: "An IP address that is the focus for the ACL",
							Required:    true,
						},
						"negated": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "A boolean that will negate the match if true",
						},
						"subnet": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An optional subnet mask applied to the IP address",
						},
					},
				},
			},
			"manage_entries": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Whether to reapply changes if the state of the entries drifts, i.e. if entries are managed externally",
			},
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Service that the ACL belongs to",
			},
		},
	}
}

func resourceServiceACLEntriesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	batchACLEntries := []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)

		entry := buildBatchACLEntry(val, gofastly.CreateBatchOperation)
		batchACLEntries = append(batchACLEntries, entry)
	}

	// Process the batch operations
	err := executeBatchACLOperations(gofastly.NewContextForResourceID(ctx, serviceID), conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclID))

	if !d.Get("manage_entries").(bool) {
		log.Print("[DEBUG] Skipping ACL entries refresh after create: manage_entries is false")
		return nil
	}

	return refreshServiceACLEntries(ctx, d, meta)
}

func resourceServiceACLEntriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if !d.Get("manage_entries").(bool) {
		log.Print("[DEBUG] Skipping ACL entries refresh: manage_entries is false (clearing entries from state)")
		if err := d.Set("entry", []map[string]any{}); err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	return refreshServiceACLEntries(ctx, d, meta)
}

func refreshServiceACLEntries(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Print("[DEBUG] Refreshing ACL Entries Configuration")

	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	remoteState, err := getAllACLEntriesViaPaginator(gofastly.NewContextForResourceID(ctx, serviceID), conn, &gofastly.GetACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("entry", flattenACLEntries(remoteState))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceACLEntriesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	if d.HasChange("manage_entries") {
		oldManageEntries, newManageEntries := d.GetChange("manage_entries")
		if !oldManageEntries.(bool) && newManageEntries.(bool) {
			return reconcileServiceACLEntriesAfterEnablingManagement(ctx, d, meta)
		}
	}

	batchACLEntries := []*gofastly.BatchACLEntry{}

	if d.HasChange("entry") {
		oe, ne := d.GetChange("entry")

		if oe == nil {
			oe = new(schema.Set)
		}
		if ne == nil {
			ne = new(schema.Set)
		}

		oldSet := oe.(*schema.Set)
		newSet := ne.(*schema.Set)

		setDiff := NewSetDiff(func(resource any) (any, error) {
			t, ok := resource.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("resource failed to be type asserted: %+v", resource)
			}
			return t["id"], nil
		})

		diffResult, err := setDiff.Diff(oldSet, newSet)
		if err != nil {
			return diag.FromErr(err)
		}

		// DELETE removed resources
		for _, resource := range diffResult.Deleted {
			resource := resource.(map[string]any)

			batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
				Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
				EntryID:   gofastly.ToPointer(resource["id"].(string)),
			})
		}

		// CREATE new resources
		for _, resource := range diffResult.Added {
			resource := resource.(map[string]any)

			entry := buildBatchACLEntry(resource, gofastly.CreateBatchOperation)
			batchACLEntries = append(batchACLEntries, entry)
		}

		// UPDATE modified resources
		for _, resource := range diffResult.Modified {
			resource := resource.(map[string]any)

			entry := buildBatchACLEntry(resource, gofastly.UpdateBatchOperation)
			batchACLEntries = append(batchACLEntries, entry)
		}
	}

	// Process the batch operations
	err := executeBatchACLOperations(gofastly.NewContextForResourceID(ctx, serviceID), conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error updating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	if !d.Get("manage_entries").(bool) {
		log.Print("[DEBUG] Skipping ACL entries refresh after update: manage_entries is false")
		return nil
	}

	return refreshServiceACLEntries(ctx, d, meta)
}

func reconcileServiceACLEntriesAfterEnablingManagement(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	desiredEntries := d.Get("entry").(*schema.Set)

	remoteState, err := getAllACLEntriesViaPaginator(gofastly.NewContextForResourceID(ctx, serviceID), conn, &gofastly.GetACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	remoteEntries := flattenACLEntries(remoteState)
	matchedRemoteEntries := make([]bool, len(remoteEntries))
	remoteEntryIndexes := make(map[aclEntryKey][]int, len(remoteEntries))
	batchACLEntries := []*gofastly.BatchACLEntry{}

	for i, remoteEntry := range remoteEntries {
		key := newACLEntryKey(remoteEntry)
		remoteEntryIndexes[key] = append(remoteEntryIndexes[key], i)
	}

	// Entries are intentionally absent from state while manage_entries is false.
	// When management is enabled again, match configured entries to the current
	// remote ACL by value so existing entries can be retained without relying on
	// state-only entry IDs.
	for _, rawDesiredEntry := range desiredEntries.List() {
		desiredEntry := rawDesiredEntry.(map[string]any)
		key := newACLEntryKey(desiredEntry)
		matchingIndexes := remoteEntryIndexes[key]

		if len(matchingIndexes) > 0 {
			matchedRemoteEntries[matchingIndexes[0]] = true
			remoteEntryIndexes[key] = matchingIndexes[1:]
			continue
		}

		batchACLEntries = append(batchACLEntries, buildBatchACLEntry(desiredEntry, gofastly.CreateBatchOperation))
	}

	// Once management is enabled, remote entries that are not represented in
	// configuration must be removed to restore the normal manage_entries=true
	// reconciliation semantics.
	deletions := []*gofastly.BatchACLEntry{}
	for i, remoteEntry := range remoteEntries {
		if matchedRemoteEntries[i] {
			continue
		}

		entryID, ok := remoteEntry["id"].(string)
		if !ok || entryID == "" {
			return diag.Errorf("error reconciling ACL entries after enabling management: remote ACL entry is missing an ID")
		}

		deletions = append(deletions, &gofastly.BatchACLEntry{
			Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
			EntryID:   gofastly.ToPointer(entryID),
		})
	}

	// Delete unmanaged remote entries before creating any missing configured
	// entries. This mirrors the ordering used by the regular update path.
	batchACLEntries = append(deletions, batchACLEntries...)

	if err := executeBatchACLOperations(gofastly.NewContextForResourceID(ctx, serviceID), conn, serviceID, aclID, batchACLEntries); err != nil {
		return diag.Errorf("error reconciling ACL entries after enabling management: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	return refreshServiceACLEntries(ctx, d, meta)
}

type aclEntryKey struct {
	ip      string
	subnet  string
	negated bool
	comment string
}

func newACLEntryKey(entry map[string]any) aclEntryKey {
	key := aclEntryKey{}

	key.ip, _ = entry["ip"].(string)
	key.subnet, _ = entry["subnet"].(string)
	key.negated, _ = entry["negated"].(bool)
	key.comment, _ = entry["comment"].(string)

	return key
}

func resourceServiceACLEntriesDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	batchACLEntries := []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]any)

		batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
			Operation: gofastly.ToPointer(gofastly.DeleteBatchOperation),
			EntryID:   gofastly.ToPointer(val["id"].(string)),
		})
	}

	// Process the batch operations
	err := executeBatchACLOperations(gofastly.NewContextForResourceID(ctx, serviceID), conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId("")
	return nil
}

// flattenACLEntries models data into format suitable for saving to Terraform state.
func flattenACLEntries(remoteState []*gofastly.ACLEntry) []map[string]any {
	var result []map[string]any

	for _, resource := range remoteState {
		data := map[string]any{}

		if resource.EntryID != nil {
			data["id"] = *resource.EntryID
		}
		if resource.IP != nil {
			data["ip"] = *resource.IP
		}
		if resource.Negated != nil {
			data["negated"] = *resource.Negated
		}
		if resource.Comment != nil {
			data["comment"] = *resource.Comment
		}
		if resource.Subnet != nil {
			data["subnet"] = strconv.Itoa(*resource.Subnet)
		}

		for k, v := range data {
			if v == "" {
				delete(data, k)
			}
		}

		result = append(result, data)
	}

	return result
}

func resourceServiceACLEntriesImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("invalid id: %s. The ID should be in the format [service_id]/[acl_id]", d.Id())
	}

	serviceID := split[0]
	aclID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("error importing ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	err = d.Set("acl_id", aclID)
	if err != nil {
		return nil, fmt.Errorf("error importing ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	err = d.Set("manage_entries", true)
	if err != nil {
		return nil, fmt.Errorf("error importing ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	return []*schema.ResourceData{d}, nil
}

func executeBatchACLOperations(ctx context.Context, conn *gofastly.Client, serviceID, aclID string, batchACLEntries []*gofastly.BatchACLEntry) error {
	batchSize := gofastly.BatchModifyMaximumOperations

	for i := 0; i < len(batchACLEntries); i += batchSize {
		j := i + batchSize
		if j > len(batchACLEntries) {
			j = len(batchACLEntries)
		}

		err := conn.BatchModifyACLEntries(ctx, &gofastly.BatchModifyACLEntriesInput{
			ServiceID: serviceID,
			ACLID:     aclID,
			Entries:   batchACLEntries[i:j],
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func buildBatchACLEntry(v map[string]any, op gofastly.BatchOperation) *gofastly.BatchACLEntry {
	entry := &gofastly.BatchACLEntry{
		Operation: gofastly.ToPointer(op),
		IP:        gofastly.ToPointer(v["ip"].(string)),
		Negated:   gofastly.ToPointer(gofastly.Compatibool(v["negated"].(bool))),
		Comment:   gofastly.ToPointer(v["comment"].(string)),
	}

	// Entry IDs are computed by Fastly and are not guaranteed to be present in
	// configuration-derived values. Create operations do not require an ID.
	if op != gofastly.CreateBatchOperation {
		if entryID, ok := v["id"].(string); ok && entryID != "" {
			entry.EntryID = gofastly.ToPointer(entryID)
		}
	}

	subnet := convertSubnetToInt(v["subnet"].(string))
	// only set zero subnet if the attribute is explicitly set
	if v["subnet"].(string) == "0" || subnet != 0 {
		entry.Subnet = gofastly.ToPointer(subnet)
	}

	return entry
}

func convertSubnetToInt(s string) int {
	subnet, _ := strconv.Atoi(s)
	return subnet
}

func getAllACLEntriesViaPaginator(ctx context.Context, conn *gofastly.Client, input *gofastly.GetACLEntriesInput) ([]*gofastly.ACLEntry, error) {
	paginator := conn.GetACLEntries(ctx, input)

	var entries []*gofastly.ACLEntry
	for paginator.HasNext() {
		results, err := paginator.GetNext()
		if err != nil {
			return nil, err
		}
		entries = append(entries, results...)
	}
	return entries, nil
}
