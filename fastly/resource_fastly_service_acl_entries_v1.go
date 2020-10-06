package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"strings"
)

func resourceServiceAclEntriesV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceAclEntriesV1Create,
		Read:   resourceServiceAclEntriesV1Read,
		Update: resourceServiceAclEntriesV1Update,
		Delete: resourceServiceAclEntriesV1Delete,
		Importer: &schema.ResourceImporter{
			State: resourceServiceACLEntriesV1Import,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Id",
			},

			"acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ACL Id",
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
							Description: "",
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

func resourceServiceAclEntriesV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	var batchACLEntries = []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]interface{})

		batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
			Operation: gofastly.CreateBatchOperation,
			IP:        val["ip"].(string),
			Subnet:    val["subnet"].(string),
			Negated:   val["negated"].(bool),
			Comment:   val["comment"].(string),
		})
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return fmt.Errorf("Error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclID))
	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)

	aclEntries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		Service: serviceID,
		ACL:     aclID,
	})

	if err != nil {
		return err
	}

	d.Set("entry", flattenAclEntries(aclEntries))
	return nil
}

func resourceServiceAclEntriesV1Update(d *schema.ResourceData, meta interface{}) error {

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

		oes := oe.(*schema.Set)
		nes := ne.(*schema.Set)

		removeEntries := oes.Difference(nes).List()
		addEntries := nes.Difference(oes).List()

		// DELETE old ACL entry
		for _, vRaw := range removeEntries {
			val := vRaw.(map[string]interface{})

			batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
				Operation: gofastly.DeleteBatchOperation,
				ID:        val["id"].(string),
			})
		}

		// POST new ACL entry
		for _, vRaw := range addEntries {
			val := vRaw.(map[string]interface{})

			batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
				Operation: gofastly.CreateBatchOperation,
				ID:        val["id"].(string),
				IP:        val["ip"].(string),
				Subnet:    val["subnet"].(string),
				Negated:   val["negated"].(bool),
				Comment:   val["comment"].(string),
			})
		}
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return fmt.Errorf("Error updating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
	}

	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	serviceID := d.Get("service_id").(string)
	aclID := d.Get("acl_id").(string)
	entries := d.Get("entry").(*schema.Set)

	var batchACLEntries = []*gofastly.BatchACLEntry{}

	for _, vRaw := range entries.List() {
		val := vRaw.(map[string]interface{})

		batchACLEntries = append(batchACLEntries, &gofastly.BatchACLEntry{
			Operation: gofastly.DeleteBatchOperation,
			ID:        val["id"].(string),
		})
	}

	// Process the batch operations
	err := executeBatchACLOperations(conn, serviceID, aclID, batchACLEntries)
	if err != nil {
		return fmt.Errorf("Error creating ACL entries: service %s, ACL %s, %s", serviceID, aclID, err)
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
			"subnet":  currentAclEntry.Subnet,
			"negated": currentAclEntry.Negated,
			"comment": currentAclEntry.Comment,
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

func resourceServiceACLEntriesV1Import(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
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
			Service: serviceID,
			ACL:     aclID,
			Entries: batchACLEntries[i:j],
		})

		if err != nil {
			return err
		}

	}

	return nil
}
