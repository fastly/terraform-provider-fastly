package fastly

import (
	"context"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceServiceAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceAuthorizationCreate,
		ReadContext:   resourceServiceAuthorizationRead,
		UpdateContext: resourceServiceAuthorizationUpdate,
		DeleteContext: resourceServiceAuthorizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of this service authorization.",
			},
			"permission": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The permissions to grant the user. Can be `full`, `read_only`, `purge_select` or `purge_all`.",
				ValidateDiagFunc: validateServiceAuthorizationPermission(),
			},
			"service_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The ID of the service to grant permissions for.",
			},

			"user_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
				Description: "The ID of the user which will receive the granted permissions.",
			},
		},
	}
}

func resourceServiceAuthorizationCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	sa, err := conn.CreateServiceAuthorization(&gofastly.CreateServiceAuthorizationInput{
		Service:    &gofastly.SAService{ID: d.Get("service_id").(string)},
		User:       &gofastly.SAUser{ID: d.Get("user_id").(string)},
		Permission: d.Get("permission").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(sa.ID)

	return nil
}

func resourceServiceAuthorizationRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing Service Authorization Configuration for (%s)", d.Id())

	conn := meta.(*APIClient).conn

	sa, err := conn.GetServiceAuthorization(&gofastly.GetServiceAuthorizationInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(sa.ID)
	d.Set("service_id", sa.Service.ID)
	d.Set("user_id", sa.User.ID)
	d.Set("permission", sa.Permission)

	return nil
}

func resourceServiceAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	if d.HasChanges("permission") {
		_, err := conn.UpdateServiceAuthorization(&gofastly.UpdateServiceAuthorizationInput{
			ID:          d.Id(),
			Permissions: d.Get("permission").(string),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceServiceAuthorizationRead(ctx, d, meta)
}

func resourceServiceAuthorizationDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteServiceAuthorization(&gofastly.DeleteServiceAuthorizationInput{
		ID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
