package fastly

import (
	"context"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

func resourceUserCreate(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	u, err := conn.CreateUser(&gofastly.CreateUserInput{
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

func resourceUserRead(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	log.Printf("[DEBUG] Refreshing User Configuration for (%s)", d.Id())
	conn := meta.(*APIClient).conn

	u, err := conn.GetUser(&gofastly.GetUserInput{
		UserID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if u.Login != nil {
		d.Set("login", u.Login)
	}
	if u.Name != nil {
		d.Set("name", u.Name)
	}
	if u.Role != nil {
		d.Set("role", u.Role)
	}

	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	// Update Name and/or Role.
	if d.HasChanges("name", "role") {
		_, err := conn.UpdateUser(&gofastly.UpdateUserInput{
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

func resourceUserDelete(_ context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	err := conn.DeleteUser(&gofastly.DeleteUserInput{
		UserID: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
