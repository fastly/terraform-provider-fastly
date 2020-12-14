package fastly

import (
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourcePrivateKeyV1() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateKeyV1Create,
		Read:   resourcePrivateKeyV1Read,
		// This resource has no update
		Delete: resourcePrivateKeyV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed
			"key_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"public_key_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePrivateKeyV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	pk, err := conn.CreatePrivateKey(&gofastly.CreatePrivateKeyInput{
		Key:  d.Get("key").(string),
		Name: d.Get("name").(string),
	})

	if err != nil {
		return err
	}

	d.SetId(pk.ID)

	return resourcePrivateKeyV1Read(d, meta)
}

func resourcePrivateKeyV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	pk, err := conn.GetPrivateKey(&gofastly.GetPrivateKeyInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	d.Set("name", pk.Name)
	d.Set("key_length", pk.KeyLength)
	d.Set("key_type", pk.KeyType)
	d.Set("replace", pk.Replace)
	d.Set("public_key_sha1", pk.PublicKeySHA1)
	d.Set("created_at", pk.CreatedAt.String())

	return nil
}

func resourcePrivateKeyV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeletePrivateKey(&gofastly.DeletePrivateKeyInput{
		ID: d.Id(),
	})

	if err != nil {
		return err
	}

	return nil
}
