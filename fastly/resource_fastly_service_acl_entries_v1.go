package fastly

import (
"fmt"
gofastly "github.com/fastly/go-fastly/fastly"
"github.com/hashicorp/terraform/helper/schema"
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
				Type: schema.TypeSet,
				Optional: true,
				Description: "ACL Entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Description: "",
							Required: true,
						},
						"subnet": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "",
						},
						"negated": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "",
						},
						"comment": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "",
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



	// TODO review this
	d.SetId(fmt.Sprintf("%s/%s", serviceID, aclId))
	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Update(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")

	return resourceServiceAclEntriesV1Read(d, meta)
}

func resourceServiceAclEntriesV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")

	return nil
}

func resourceServiceAclEntriesV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn
	comp := strings.Split(d.Id(), "/")



	d.SetId("")
	return nil
}

func flattenAclEntries(aclEntryList []*gofastly.ACLEntry) []map[string]interface{} {

	var resultList []map[string]interface{}

	for _, currentAclEntry := range aclEntryList {
		aes := map[string]interface{}{
			"ip": currentAclEntry.IP,
			"subnet": currentAclEntry.Subnet,
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