package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSPrivateKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceTLSPrivateKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Fastly private key ID. Conflicts with all the other filters",
				ConflictsWith: []string{"name", "created_at", "key_length", "key_type", "public_key_sha1"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The human-readable name assigned to the private key when uploaded.",
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
				Description:   "A hash of the associated public key, useful for safely identifying it.",
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
		filters := getTLSPrivateKeyFilters(d)

		privateKeys, err := listTLSPrivateKeys(conn, filters...)
		if err != nil {
			return err
		}

		if len(privateKeys) == 0 {
			return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
		}

		if len(privateKeys) > 1 {
			return fmt.Errorf("Your query returned more than one result. Please change to a more specific search criteria and try again.")
		}

		privateKey = privateKeys[0]
	}

	return dataSourceFastlyTLSPrivateKeySetAttributes(privateKey, d)
}

type TLSPrivateKeyPredicate func(key *fastly.PrivateKey) bool

func getTLSPrivateKeyFilters(d *schema.ResourceData) []TLSPrivateKeyPredicate {
	var filters []TLSPrivateKeyPredicate

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

	return filters
}

func listTLSPrivateKeys(conn *fastly.Client, filters ...TLSPrivateKeyPredicate) ([]*fastly.PrivateKey, error) {
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

func dataSourceFastlyTLSPrivateKeySetAttributes(privateKey *fastly.PrivateKey, d *schema.ResourceData) error {
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

func filterPrivateKey(privateKey *fastly.PrivateKey, filters []TLSPrivateKeyPredicate) bool {
	for _, f := range filters {
		if !f(privateKey) {
			return false
		}
	}
	return true
}
