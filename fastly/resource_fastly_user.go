package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"login": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The email address, which is the login name, of the User",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The real life name of the user",
			},

			"role": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "user",
				Description:      "The role of this user. Can be `user` (the default), `billing`, `engineer`, or `superuser`. For detailed information on the abilities granted to each role, see [Fastly's Documentation on User roles](https://docs.fastly.com/en/guides/configuring-user-roles-and-permissions#user-roles-and-what-they-can-do)",
				ValidateDiagFunc: validateUserRole(),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	u, err := conn.CreateUser(ctx, &gofastly.CreateUserInput{
		Login: gofastly.ToPointer(d.Get("login").(string)),
		Name:  gofastly.ToPointer(d.Get("name").(string)),
		Role:  gofastly.ToPointer(d.Get("role").(string)),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if u.UserID == nil {
		return diag.Errorf("error: user.ID is nil")
	}
	d.SetId(*u.UserID)

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing User Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	u, err := conn.GetUser(ctx, &gofastly.GetUserInput{
		UserID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if u.Login != nil {
		err = d.Set("login", u.Login)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if u.Name != nil {
		err = d.Set("name", u.Name)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if u.Role != nil {
		err = d.Set("role", u.Role)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// Update Name and/or Role.
	if d.HasChanges("name", "role") {
		_, err := conn.UpdateUser(ctx, &gofastly.UpdateUserInput{
			UserID: d.Id(),
			Name:   gofastly.ToPointer(d.Get("name").(string)),
			Role:   gofastly.ToPointer(d.Get("role").(string)),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteUser(ctx, &gofastly.DeleteUserInput{
		UserID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
