package fastly

import (
	"fmt"
	"log"
	"strings"
	"time"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFastlyConfigStoreEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceFastlyConfigStoreEntryCreate,
		Read:   resourceFastlyConfigStoreEntryRead,
		Update: resourceFastlyConfigStoreEntryUpdate,
		Delete: resourceFastlyConfigStoreEntryDelete,
		Importer: &schema.ResourceImporter{
			State: resourceFastlyConfigStoreEntryImport,
		},
		Description: "Manages an individual entry in a Fastly ConfigStore",
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Key of the ConfigStore entry",
			},
			"store_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the ConfigStore",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value of the ConfigStore entry",
			},
		},
	}
}

func resourceFastlyConfigStoreEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*APIClient).conn
	storeID := d.Get("store_id").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	log.Printf("[DEBUG] Creating ConfigStore entry: %s/%s", storeID, key)

	_, err := conn.CreateConfigStoreItem(&gofastly.CreateConfigStoreItemInput{
		StoreID: storeID,
		Key:     key,
		Value:   value,
	})
	if err != nil {
		return fmt.Errorf("error creating ConfigStore entry (%s/%s): %s", storeID, key, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", storeID, key))
	return resourceFastlyConfigStoreEntryRead(d, meta)
}

func resourceFastlyConfigStoreEntryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*APIClient).conn

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return fmt.Errorf("invalid resource ID: %s (expected format: store_id/key)", d.Id())
	}
	storeID := idParts[0]
	key := idParts[1]

	log.Printf("[DEBUG] Reading ConfigStore entry: %s/%s", storeID, key)

	item, err := conn.GetConfigStoreItem(&gofastly.GetConfigStoreItemInput{
		StoreID: storeID,
		Key:     key,
	})
	if err != nil {
		if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
			log.Printf("[WARN] ConfigStore entry (%s/%s) not found, removing from state", storeID, key)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error looking up ConfigStore entry (%s/%s): %s", storeID, key, err)
	}

	d.Set("store_id", storeID)
	d.Set("key", item.Key)
	d.Set("value", item.Value)

	return nil
}

func resourceFastlyConfigStoreEntryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*APIClient).conn

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return fmt.Errorf("invalid resource ID: %s (expected format: store_id/key)", d.Id())
	}
	storeID := idParts[0]
	key := idParts[1]
	value := d.Get("value").(string)

	log.Printf("[DEBUG] Updating ConfigStore entry: %s/%s", storeID, key)

	_, err := conn.UpdateConfigStoreItem(&gofastly.UpdateConfigStoreItemInput{
		StoreID: storeID,
		Key:     key,
		Value:   value,
	})
	if err != nil {
		return fmt.Errorf("error updating ConfigStore entry (%s/%s): %s", storeID, key, err)
	}

	return resourceFastlyConfigStoreEntryRead(d, meta)
}

func resourceFastlyConfigStoreEntryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*APIClient).conn

	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return fmt.Errorf("invalid resource ID: %s (expected format: store_id/key)", d.Id())
	}
	storeID := idParts[0]
	key := idParts[1]

	log.Printf("[DEBUG] Deleting ConfigStore entry: %s/%s", storeID, key)

	err := conn.DeleteConfigStoreItem(&gofastly.DeleteConfigStoreItemInput{
		StoreID: storeID,
		Key:     key,
	})
	if err != nil {
		return fmt.Errorf("error deleting ConfigStore entry (%s/%s): %s", storeID, key, err)
	}

	return nil
}

func resourceFastlyConfigStoreEntryImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("invalid import ID, expected format: store_id/key (got: %s)", d.Id())
	}

	storeID := idParts[0]
	key := idParts[1]

	d.Set("store_id", storeID)
	d.Set("key", key)
	d.SetId(fmt.Sprintf("%s/%s", storeID, key))

	return []*schema.ResourceData{d}, resourceFastlyConfigStoreEntryRead(d, meta)
}