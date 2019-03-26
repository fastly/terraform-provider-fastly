package fastly

import (
	"github.com/hashicorp/terraform/helper/schema"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

func resourceACLEntryV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceACLEntryV1Create,
		Read:   resourceACLEntryV1Read,
		Update: resourceACLEntryV1Update,
		Delete: resourceACLEntryV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the service",
			},
			"acl_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the ACL",
			},
			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address",
			},
			"subnet": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Subnet",
			},
			"negated": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Negated",
			},
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comment",
			},
		},
	}
}

func resourceACLEntryV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	aclEntry, err := conn.CreateACLEntry(&gofastly.CreateACLEntryInput{
		Service: d.Get("service_id").(string),
		ACL:     d.Get("acl_id").(string),
		IP:      d.Get("ip").(string),
		Subnet:  d.Get("subnet").(string),
		Negated: d.Get("negated").(bool),
		Comment: d.Get("comment").(string),
	})

	if err != nil {
		return err
	}

	d.SetId(aclEntry.ID)
	return resourceACLEntryV1Read(d, meta)
}

func resourceACLEntryV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	_, err := conn.UpdateACLEntry(&gofastly.UpdateACLEntryInput{
		Service: d.Get("service_id").(string),
		ACL:     d.Get("acl_id").(string),
		ID:      d.Id(),
		IP:      d.Get("ip").(string),
		Subnet:  d.Get("subnet").(string),
		Negated: d.Get("negated").(bool),
		Comment: d.Get("comment").(string),
	})

	if err != nil {
		return err
	}

	return resourceACLEntryV1Read(d, meta)
}

func resourceACLEntryV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	aclEntry, err := conn.GetACLEntry(&gofastly.GetACLEntryInput{
		Service: d.Get("service_id").(string),
		ACL:     d.Get("acl_id").(string),
		ID:      d.Id(),
	})

	if err != nil {
		return err
	}

	d.SetId(aclEntry.ID)
	d.Set("service_id", aclEntry.ServiceID)
	d.Set("acl_id", aclEntry.ACLID)
	d.Set("ip", aclEntry.IP)
	d.Set("subnet", aclEntry.Subnet)
	d.Set("negated", aclEntry.Negated)
	d.Set("comment", aclEntry.Comment)

	return nil
}

func resourceACLEntryV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteACLEntry(&gofastly.DeleteACLEntryInput{
		Service: d.Get("service_id").(string),
		ACL:     d.Get("acl_id").(string),
		ID:      d.Id(),
	})

	if err != nil {
		return err
	}

	return nil
}
