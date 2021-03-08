package fastly

import (
	"context"
	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserV1Create,
		ReadContext:   resourceUserV1Read,
		UpdateContext: resourceUserV1Update,
		DeleteContext: resourceUserV1Delete,
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

func resourceUserV1Create(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	u, err := conn.CreateUser(&gofastly.CreateUserInput{
		Login: d.Get("login").(string),
		Name:  d.Get("name").(string),
		Role:  d.Get("role").(string),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(u.ID)

	return nil
}

func resourceUserV1Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	u, err := conn.GetUser(&gofastly.GetUserInput{
		ID: d.Id(),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("login", u.Login)
	d.Set("name", u.Name)
	d.Set("role", u.Role)

	return nil
}

func resourceUserV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	// Update Name and/or Role.
	if d.HasChanges("name", "role") {
		_, err := conn.UpdateUser(&gofastly.UpdateUserInput{
			ID:   d.Id(),
			Name: gofastly.String(d.Get("name").(string)),
			Role: gofastly.String(d.Get("role").(string)),
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceUserV1Read(ctx, d, meta)
}

func resourceUserV1Delete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteUser(&gofastly.DeleteUserInput{
		ID: d.Id(),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
