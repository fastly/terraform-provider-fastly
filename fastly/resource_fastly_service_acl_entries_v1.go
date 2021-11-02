package fastly

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	gofastly "github.com/fastly/go-fastly/v5/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceAclEntriesV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceAclEntriesV1Create,
		ReadContext:   resourceServiceAclEntriesV1Read,
		UpdateContext: resourceServiceAclEntriesV1Update,
		DeleteContext: resourceServiceAclEntriesV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceACLEntriesV1Import,
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
			"entry": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "ACL Entries",
				MaxItems:    gofastly.MaximumACLSize,
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

func resourceServiceAclEntriesV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	var batchACLEntries = []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]interface{})

		entry := buildBatchACLEntry(val, gofastly.CreateBatchOperation)
		batchACLEntries = append(batchACLEntries, entry)
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return diag.Errorf("Error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclID))
	return resourceServiceAclEntriesV1Read(ctx, d, meta)
}

func resourceServiceAclEntriesV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	aclEntries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		ServiceID: serviceID,
		ACLID:     aclID,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("entry", flattenAclEntries(aclEntries))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceAclEntriesV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	var batchACLEntries = []*gofastly.BatchACLEntry{}

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
		return diag.Errorf("Error updating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	return resourceServiceAclEntriesV1Read(ctx, d, meta)
}

func resourceServiceAclEntriesV1Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	var batchACLEntries = []*gofastly.BatchACLEntry{}

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
		return diag.Errorf("Error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId("")
	return nil
}

func flattenAclEntries(aclEntryList []*gofastly.ACLEntry) []map[string]interface{} {

	var resultList []map[string]interface{}

	for _, currentAclEntry := range aclEntryList {
		aes := map[string]interface{}{
			"id":      currentAclEntry.ID,
			"ip":      currentAclEntry.IP,
			"negated": currentAclEntry.Negated,
			"comment": currentAclEntry.Comment,
		}

		// NOTE: Fastly API may return "null" or int value
		// we only want to set the value if subnet is not null
		if currentAclEntry.Subnet != nil {
			aes["subnet"] = strconv.Itoa(*currentAclEntry.Subnet)
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

func resourceServiceACLEntriesV1Import(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	split := strings.Split(d.Id(), "/")

	if len(split) != 2 {
		return nil, fmt.Errorf("Invalid id: %s. The ID should be in the format [service_id]/[acl_id]", d.Id())
	}

	serviceID := split[0]
	aclID := split[1]

	err := d.Set("service_id", serviceID)
	if err != nil {
		return nil, fmt.Errorf("Error importing ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	err = d.Set("acl_id", aclID)
	if err != nil {
		return nil, fmt.Errorf("Error importing ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
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
	cond1 := isZeroSubnet(v["ip"].(string)) && v["subnet"].(string) == "0"
	cond2 := subnet != 0
	if cond1 || cond2 {
		entry.Subnet = gofastly.Int(subnet)
	}

	return entry
}

func convertSubnetToInt(s string) int {
	subnet, _ := strconv.Atoi(s)
	return subnet
}

func isZeroSubnet(ip string) bool {
	return ip == "0.0.0.0" || ip == "::"
}
