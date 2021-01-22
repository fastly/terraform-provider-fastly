package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceTLSPrivateKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTLSPrivateKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Fastly private key ID",
				ConflictsWith: []string{"name", "created_at", "key_length", "key_type", "public_key_sha1"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Customisable name of the private key.",
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"created_at": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Timestamp (GMT) when the private key was created.",
				ConflictsWith: []string{"id"},
			},
			"key_length": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				Description:   "The key length used to generate the private key.",
				ConflictsWith: []string{"id"},
			},
			"key_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The algorithm used to generate the private key. Must be RSA.",
				ConflictsWith: []string{"id"},
			},
			"replace": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Fastly recommends replacing this private key.",
			},
			"public_key_sha1": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Useful for safely identifying the key.",
				ConflictsWith: []string{"id"},
			},
		},
	}
}

func dataSourceTLSPrivateKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var privateKey *fastly.PrivateKey
	if v, ok := d.GetOk("id"); ok {
		var err error
		privateKey, err = conn.GetPrivateKey(&fastly.GetPrivateKeyInput{ID: v.(string)})
		if err != nil {
			return err
		}
	} else {
		var err error
		privateKey, err = findTLSPrivateKey(conn, d)
		if err != nil {
			return err
		}
	}

	d.SetId(privateKey.ID)
	if err := d.Set("name", privateKey.Name); err != nil {
		return err
	}
	if err := d.Set("created_at", privateKey.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("key_length", privateKey.KeyLength); err != nil {
		return err
	}
	if err := d.Set("key_type", privateKey.KeyType); err != nil {
		return err
	}
	if err := d.Set("replace", privateKey.Replace); err != nil {
		return err
	}
	if err := d.Set("public_key_sha1", privateKey.PublicKeySHA1); err != nil {
		return err
	}
	return nil
}

func findTLSPrivateKey(conn *fastly.Client, d *schema.ResourceData) (*fastly.PrivateKey, error) {
	var filters []func(*fastly.PrivateKey) bool

	if v, ok := d.GetOk("id"); ok {
		filters = append(filters, func(key *fastly.PrivateKey) bool {
			return key.ID == v.(string)
		})
	}
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

	privateKeys, err := listTLSPrivateKeys(conn, filters...)
	if err != nil {
		return nil, err
	}

	if len(privateKeys) == 0 {
		return nil, fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(privateKeys) > 1 {
		return nil, fmt.Errorf("Your query returned more than one result. Please change to a more specific search criteria and try again.")
	}

	return privateKeys[0], nil
}

func listTLSPrivateKeys(conn *fastly.Client, filters ...func(*fastly.PrivateKey) bool) ([]*fastly.PrivateKey, error) {
	var privateKeys []*fastly.PrivateKey
	pageNumber := 1
	for {
		list, err := conn.ListPrivateKeys(&fastly.ListPrivateKeysInput{
			PageNumber: pageNumber,
		})
		if err != nil {
			return nil, err
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
	return privateKeys, nil
}

func filterPrivateKey(privateKey *fastly.PrivateKey, filters []func(*fastly.PrivateKey) bool) bool {
	for _, f := range filters {
		if !f(privateKey) {
			return false
		}
	}
	return true
}
