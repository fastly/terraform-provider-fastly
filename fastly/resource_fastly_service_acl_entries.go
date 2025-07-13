package fastly

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v11/fastly"
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
	return resourceServiceACLEntriesRead(ctx, d, meta)
}

func resourceServiceACLEntriesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	return resourceServiceACLEntriesRead(ctx, d, meta)
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
		EntryID:   gofastly.ToPointer(v["id"].(string)),
		IP:        gofastly.ToPointer(v["ip"].(string)),
		Negated:   gofastly.ToPointer(gofastly.Compatibool(v["negated"].(bool))),
		Comment:   gofastly.ToPointer(v["comment"].(string)),
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
