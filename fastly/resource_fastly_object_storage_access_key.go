package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/objectstorage/accesskeys"
)

func resourceObjectStorageAccessKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceObjectStorageAccessKeyCreate,
		ReadContext:   resourceObjectStorageAccessKeyRead,
		DeleteContext: resourceObjectStorageAccessKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"access_key_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID for the object storage access token",
			},
			"buckets": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Optional list of buckets the access key will be associated with.  Example: `[\"bucket1\", \"bucket2\"]`",
				Elem:        &schema.Schema{Type: schema.TypeString},
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the access key",
				ForceNew:    true,
			},
			"permission": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The permissions of the access key",
				ForceNew:    true,
			},
			"secret_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   !DisplaySensitiveFields,
				Description: "Secret key for the object storage access token",
			},
		},
	}
}

func resourceObjectStorageAccessKeyCreate(_ context.Context, resourceData *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	opts := accesskeys.CreateInput{
		Description: gofastly.ToPointer(resourceData.Get("description").(string)),
		Permission:  gofastly.ToPointer(resourceData.Get("permission").(string)),
	}

	buckets := []string{}
	if val, ok := resourceData.GetOk("buckets"); ok {
		for _, bucket := range val.([]any) {
			buckets = append(buckets, bucket.(string))
		}
		opts.Buckets = gofastly.ToPointer(buckets)
	}
	createdAK, err := accesskeys.Create(conn, &opts)
	if err != nil {
		return diag.FromErr(err)
	}

	if createdAK.AccessKeyID == "" {
		return diag.Errorf("error: accessKey.AccessKeyID is empty")
	}
	resourceData.SetId(createdAK.AccessKeyID)
	err = resourceData.Set("access_key_id", createdAK.AccessKeyID)
	if err != nil {
		return diag.FromErr(err)
	}

	if createdAK.SecretKey == "" {
		return diag.Errorf("error: accessKey.SecretKey is empty")
	}
	err = resourceData.Set("secret_key", createdAK.SecretKey)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceObjectStorageAccessKeyRead(_ context.Context, resourceData *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Object Storage Access Key Configuration for (%s)", resourceData.Get("access_key_id"))
	conn := meta.(*APIClient).conn

	opts := accesskeys.GetInput{
		AccessKeyID: gofastly.ToPointer(resourceData.Id()),
	}

	readAK, err := accesskeys.Get(conn, &opts)
	if err != nil {
		return diag.FromErr(err)
	}
	if readAK.Permission != "" {
		err = resourceData.Set("permission", readAK.Permission)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if readAK.Description != "" {
		err = resourceData.Set("description", readAK.Description)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if len(readAK.Buckets) != 0 {
		err = resourceData.Set("buckets", readAK.Buckets)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if readAK.SecretKey != "" {
		err = resourceData.Set("secret_key", readAK.SecretKey)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if readAK.AccessKeyID != "" {
		err = resourceData.Set("access_key_id", readAK.SecretKey)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceObjectStorageAccessKeyDelete(_ context.Context, resourceData *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	opts := accesskeys.DeleteInput{
		AccessKeyID: gofastly.ToPointer(resourceData.Id()),
	}

	err := accesskeys.Delete(conn, &opts)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
