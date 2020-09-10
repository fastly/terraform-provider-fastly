package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

type dataSourceFastlyACLEntriesResult struct {
	ACLEntries []map[string]interface{}
}

func dataSourceFastlyACLEntries() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyACLEntriesRead,

		Schema: map[string]*schema.Schema{
			"service": {
				Type:     schema.TypeString,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"acl": {
				Type:     schema.TypeString,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"acl_entries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"acl_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"negated": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"comment": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyACLEntriesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	log.Printf("[DEBUG] Reading ACL Entries")

	aclEntriesInput := &fastly.ListACLEntriesInput{
		Service: d.Get("service").(string),
		ACL:     d.Get("acl").(string),
	}
	aclEntries, err := conn.ListACLEntries(aclEntriesInput)

	if err != nil {
		return fmt.Errorf("Error listing ACL entries for service: %s", err)
	}

	aclEntriesList := []map[string]interface{}{}
	for _, entry := range aclEntries {
		aclEntriesList = append(aclEntriesList, map[string]interface{}{
			"service_id": entry.ServiceID,
			"acl_id":     entry.ACLID,
			"id":         entry.ID,
			"ip":         entry.IP,
			"subnet":     entry.Subnet,
			"negated":    entry.Negated,
			"comment":    entry.Comment,
		})
	}

	d.SetId(fmt.Sprintf("%s%s", d.Get("service"), d.Get("acl")))

	log.Printf("[DEBUG] ACL entries list: %v", aclEntriesList)
	if err := d.Set("acl_entries", aclEntriesList); err != nil {
		return fmt.Errorf("Error setting acl entries: %s", err)
	}

	return nil
}
