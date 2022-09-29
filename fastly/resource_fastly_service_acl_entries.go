package fastly

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Service that the ACL belongs to",
			},

			"acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the ACL that the items belong to",
			},
			"manage_entries": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Whether to reapply changes if the state of the entries drifts, i.e. if entries are managed externally",
			},
			"entry": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "ACL Entries",
				MaxItems:    gofastly.MaximumACLSize,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.HasChange("acl_id") && !d.Get("manage_entries").(bool)
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"subnet": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An optional subnet mask applied to the IP address",
						},
						"negated": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "A boolean that will negate the match if true",
						},
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A personal freeform descriptive note",
						},
					},
				},
			},
		},
	}
}

func resourceServiceACLEntriesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	batchACLEntries := []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]interface{})

		entry := buildBatchACLEntry(val, gofastly.CreateBatchOperation)
		batchACLEntries = append(batchACLEntries, entry)
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclID))
	return resourceServiceACLEntriesRead(ctx, d, meta)
}

func resourceServiceACLEntriesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Refreshing ACL Entries Configuration")

	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	aclEntries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("entry", flattenACLEntries(aclEntries))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceACLEntriesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

		setDiff := NewSetDiff(func(resource interface{}) (interface{}, error) {
			t, ok := resource.(map[string]interface{})
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
			resource := resource.(map[string]interface{})

			batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
				Operation: gofastly.DeleteBatchOperation,
				ID:        gofastly.String(resource["id"].(string)),
			})
		}

		// CREATE new resources
		for _, resource := range diffResult.Added {
			resource := resource.(map[string]interface{})

			entry := buildBatchACLEntry(resource, gofastly.CreateBatchOperation)
			batchACLEntries = append(batchACLEntries, entry)
		}

		// UPDATE modified resources
		for _, resource := range diffResult.Modified {
			resource := resource.(map[string]interface{})

			entry := buildBatchACLEntry(resource, gofastly.UpdateBatchOperation)
			batchACLEntries = append(batchACLEntries, entry)
		}
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error updating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	return resourceServiceACLEntriesRead(ctx, d, meta)
}

func resourceServiceACLEntriesDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	batchACLEntries := []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]interface{})

		batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
			Operation: gofastly.DeleteBatchOperation,
			ID:        gofastly.String(val["id"].(string)),
		})
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId("")
	return nil
}

func flattenACLEntries(aclEntryList []*gofastly.ACLEntry) []map[string]interface{} {
	var resultList []map[string]interface{}

	for _, currentACLEntry := range aclEntryList {
		aes := map[string]interface{}{
			"id":      currentACLEntry.ID,
			"ip":      currentACLEntry.IP,
			"negated": currentACLEntry.Negated,
			"comment": currentACLEntry.Comment,
		}

		// NOTE: Fastly API may return "null" or int value
		// we only want to set the value if subnet is not null
		if currentACLEntry.Subnet != nil {
			aes["subnet"] = strconv.Itoa(*currentACLEntry.Subnet)
		}

		for k, v := range aes {
			if v == "" {
				delete(aes, k)
			}
		}

		resultList = append(resultList, aes)
	}

	return resultList
}

func resourceServiceACLEntriesImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
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

func executeBatchACLOperations(conn *gofastly.Client, serviceID, aclID string, batchACLEntries []*gofastly.BatchACLEntry) error {
	batchSize := gofastly.BatchModifyMaximumOperations

	for i := 0; i < len(batchACLEntries); i += batchSize {
		j := i + batchSize
		if j > len(batchACLEntries) {
			j = len(batchACLEntries)
		}

		err := conn.BatchModifyACLEntries(&gofastly.BatchModifyACLEntriesInput{
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

func buildBatchACLEntry(v map[string]interface{}, op gofastly.BatchOperation) *gofastly.BatchACLEntry {
	entry := &gofastly.BatchACLEntry{
		Operation: op,
		ID:        gofastly.String(v["id"].(string)),
		IP:        gofastly.String(v["ip"].(string)),
		Negated:   gofastly.CBool(v["negated"].(bool)),
		Comment:   gofastly.String(v["comment"].(string)),
	}

	subnet := convertSubnetToInt(v["subnet"].(string))
	// only set zero subnet if the attribute is explicitly set
	if v["subnet"].(string) == "0" || subnet != 0 {
		entry.Subnet = gofastly.Int(subnet)
	}

	return entry
}

func convertSubnetToInt(s string) int {
	subnet, _ := strconv.Atoi(s)
	return subnet
}
