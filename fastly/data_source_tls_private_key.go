package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func dataSourceTLSPrivateKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTLSPrivateKeyRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customisable name of the private key.",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Timestamp (GMT) when the private key was created.",
			},
			"key_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The key length used to generate the private key.",
			},
			"key_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The algorithm used to generate the private key. Must be RSA.",
			},
			"replace": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether Fastly recommends replacing this private key.",
			},
			"public_key_sha1": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Useful for safely identifying the key.",
			},
		},
	}
}

func dataSourceTLSPrivateKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var filters []func(*fastly.PrivateKey) bool

	if v, ok := d.GetOk("name"); ok {
		filters = append(filters, func(key *fastly.PrivateKey) bool {
			return key.Name == v.(string)
		})
	}
	if v, ok := d.GetOk("key_length"); ok {
		filters = append(filters, func(key *fastly.PrivateKey) bool {
			return key.KeyLength == v.(int)
		})
	}
	if v, ok := d.GetOk("key_type"); ok {
		filters = append(filters, func(key *fastly.PrivateKey) bool {
			return key.KeyType == v.(string)
		})
	}
	if v, ok := d.GetOk("public_key_sha1"); ok {
		filters = append(filters, func(key *fastly.PrivateKey) bool {
			return key.PublicKeySHA1 == v.(string)
		})
	}

	var privateKeys []*fastly.PrivateKey
	pageNumber := 1
	for {
		list, err := conn.ListPrivateKeys(&fastly.ListPrivateKeysInput{
			PageNumber: pageNumber,
		})
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, privateKey := range list {
			if filterPrivateKey(privateKey, filters) {
				privateKeys = append(privateKeys, privateKey)
			}
		}
	}

	if len(privateKeys) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(privateKeys) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please change try a more specific search criteria and try again.")
	}

	privateKey := privateKeys[0]

	d.SetId(privateKey.ID)
	err := d.Set("name", privateKey.Name)
	if err != nil {
		return err
	}
	err = d.Set("created_at", privateKey.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("key_length", privateKey.KeyLength)
	if err != nil {
		return err
	}
	err = d.Set("key_type", privateKey.KeyType)
	if err != nil {
		return err
	}
	err = d.Set("replace", privateKey.Replace)
	if err != nil {
		return err
	}
	err = d.Set("public_key_sha1", privateKey.PublicKeySHA1)
	if err != nil {
		return err
	}
	return nil
}

func filterPrivateKey(privateKey *fastly.PrivateKey, filters []func(*fastly.PrivateKey) bool) bool {
	for _, f := range filters {
		if !f(privateKey) {
			return false
		}
	}
	return true
}
