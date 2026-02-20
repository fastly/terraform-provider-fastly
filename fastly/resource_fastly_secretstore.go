package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

func resourceFastlySecretStore() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFastlySecretStoreCreate,
		ReadContext:   resourceFastlySecretStoreRead,
		DeleteContext: resourceFastlySecretStoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A human-readable name for the Secret Store. The value must contain only letters, numbers, dashes (-), underscores (_), or periods (.). It is important to note that changing this attribute will delete and recreate the Secret Store, and discard the current entries. You MUST first delete the associated resource_link block from your service before modifying this field.",
				ForceNew:    true,
			},
		},
	}
}

func resourceFastlySecretStoreCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.CreateSecretStoreInput{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] CREATE: Secret Store input: %#v", input)

	store, err := conn.CreateSecretStore(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	// NOTE: `id` is exposed as a read-only attribute.
	d.SetId(store.StoreID)

	return nil
}

func resourceFastlySecretStoreRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := &gofastly.GetSecretStoreInput{
		StoreID: d.Id(),
	}

	log.Printf("[DEBUG] REFRESH: Secret Store input: %#v", input)

	store, err := conn.GetSecretStore(ctx, input)
	if err != nil {
		log.Printf("[WARN] No Secret Store found '%s'", d.Id())
		d.SetId("")
		return diag.FromErr(err)
	}

	err = d.Set("name", store.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFastlySecretStoreDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	input := gofastly.DeleteSecretStoreInput{
		StoreID: d.Id(),
	}

	log.Printf("[DEBUG] DELETE: Secret Store input: %#v", input)

	err := conn.DeleteSecretStore(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
