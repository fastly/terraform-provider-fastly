package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"strings"
)

func resourceServiceAclEntriesV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceAclEntriesV1Create,
		Read:   resourceServiceAclEntriesV1Read,
		Update: resourceServiceAclEntriesV1Update,
		Delete: resourceServiceAclEntriesV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
	aclId := d.Get("acl_id").(string)

	entries := d.Get("entry")
	acles := entries.(*schema.Set)

	for _, vRaw := range acles.List() {
		val := vRaw.(map[string]interface{})

		opts := &gofastly.CreateACLEntryInput{
			Service: serviceID,
			ACL:     aclId,
			IP:      val["ip"].(string),
			Subnet:  val["subnet"].(string),
			Negated: val["negated"].(bool),
			Comment: val["comment"].(string),
		}

		log.Printf("[DEBUG] Create ACL Entry Opts: %#v", opts)

		_, err := conn.CreateACLEntry(opts)

		if err != nil {
			return err
		}
	}

	// TODO review this
	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclId))
	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")

	aclEntries, err := conn.ListACLEntries(&gofastly.ListACLEntriesInput{
		Service: comp[0],
		ACL:     comp[1],
	})

	filteredAclEntries := filterAclEntries(aclEntries, func(currentAclEntry gofastly.ACLEntry) bool {

		data := d.Get("entry")
		entries := data.(*schema.Set)

		for _, entry := range entries.List() {
			aclEntry := entry.(map[string]interface{})

			if aclEntry["ip"] == currentAclEntry.IP {
				return true
			}
		}

		return false
	})

	if err != nil {
		return err
	}

	acles := flattenAclEntries(filteredAclEntries)

	if err := d.Set("entry", acles); err != nil {
		log.Printf("[WARn] Error setting entry for (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceServiceAclEntriesV1Update(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")

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
			opts := gofastly.DeleteACLEntryInput{
				Service: comp[0],
				ACL:     comp[1],
				ID:      val["id"].(string),
			}

			log.Printf("[DEBUG] Fastly ACL Entry removal opts: %#v", opts)
			err := conn.DeleteACLEntry(&opts)
			if errRes, ok := err.(*gofastly.HTTPError); ok {
				if errRes.StatusCode != 404 {
					return err
				}
			} else if err != nil {
				return err
			}
		}

		// POST new ACL entry
		for _, vRaw := range addEntries {
			val := vRaw.(map[string]interface{})
			opts := gofastly.CreateACLEntryInput{
				Service: comp[0],
				ACL:     comp[1],
				IP:      val["ip"].(string),
				Subnet:  val["subnet"].(string),
				Negated: val["negated"].(bool),
				Comment: val["comment"].(string),
			}

			log.Printf("[DEBUG] Fastly ACL Entry creation opts: %#v", opts)
			_, err := conn.CreateACLEntry(&opts)
			if err != nil {
				return err
			}
		}
	}

	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")

	entries := d.Get("entry")
	acles := entries.(*schema.Set)

	for _, vRaw := range acles.List() {
		val := vRaw.(map[string]interface{})

		opts := &gofastly.DeleteACLEntryInput{
			Service: comp[0],
			ACL:     comp[1],
			ID:      val["id"].(string),
		}

		log.Printf("[DEBUG] Create ACL Entry Opts: %#v", opts)

		err := conn.DeleteACLEntry(opts)

		if err != nil {
			return err
		}
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

func filterAclEntries(aclEntries []*gofastly.ACLEntry, f func(entry gofastly.ACLEntry) bool) []*gofastly.ACLEntry {
	filteredAclEntries := make([]*gofastly.ACLEntry, 0)
	for _, entry := range aclEntries {
		if f(*entry) {
			filteredAclEntries = append(filteredAclEntries, entry)
		}
	}

	return filteredAclEntries
}
