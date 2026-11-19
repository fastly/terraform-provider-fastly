package fastly

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	gofastly "github.com/fastly/go-fastly/v19/fastly"
	bmws "github.com/fastly/go-fastly/v19/fastly/botmanagement/v1/workspaces"
)

func dataSourceFastlyBotManagementWorkspace() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to look up a Fastly Bot Management workspace, either by the ID of a service attached to it or by its workspace ID. Exactly one of `service_id` or `workspace_id` must be set.",
		ReadContext: dataSourceFastlyBotManagementWorkspaceRead,
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the Bot Management workspace was created, in RFC 3339 format.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of the Bot Management workspace.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the Bot Management workspace.",
			},
			"protection_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The protection mode of the Bot Management workspace. One of `off`, `log`, or `block`.",
			},
			"service_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"service_id", "workspace_id"},
				Description:  "The ID of a service attached to the Bot Management workspace. Conflicts with `workspace_id`.",
			},
			"services": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The IDs of the services attached to the Bot Management workspace.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the Bot Management workspace was last updated, in RFC 3339 format.",
			},
			"workspace_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"service_id", "workspace_id"},
				Description:  "The ID of the Bot Management workspace. Conflicts with `service_id`.",
			},
		},
	}
}

func dataSourceFastlyBotManagementWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*APIClient).conn

	var workspace *bmws.Workspace
	if v, ok := d.GetOk("workspace_id"); ok {
		workspaceID := v.(string)
		log.Printf("[DEBUG] Reading Bot Management workspace %s", workspaceID)

		w, err := bmws.Get(ctx, conn, &bmws.GetInput{WorkspaceID: new(workspaceID)})
		if err != nil {
			if e, ok := err.(*gofastly.HTTPError); ok && e.IsNotFound() {
				return diag.Errorf("no Bot Management workspace exists with ID %q", workspaceID)
			}
			return diag.Errorf("error reading Bot Management workspace %s: %s", workspaceID, err)
		}
		workspace = w
	} else {
		serviceID := d.Get("service_id").(string)
		log.Printf("[DEBUG] Looking up Bot Management workspace for service %s", serviceID)

		workspaces, err := bmws.List(ctx, conn, &bmws.ListInput{ServiceID: new(serviceID)})
		if err != nil {
			return diag.Errorf("error listing Bot Management workspaces: %s", err)
		}
		switch len(workspaces) {
		case 0:
			return diag.Errorf("service %q is not attached to a Bot Management workspace. A workspace is created the first time `contentguard` is set to `on` in the service's `product_enablement` `bot_management` block. Set it in an earlier apply before looking up the workspace", serviceID)
		case 1:
			workspace = &workspaces[0]
		default:
			return diag.Errorf("expected service %q to be attached to one Bot Management workspace, found %d", serviceID, len(workspaces))
		}
	}

	d.SetId(workspace.ID)

	services := workspace.Services
	if services == nil {
		services = []string{}
	}

	for k, v := range map[string]any{
		"workspace_id":    workspace.ID,
		"name":            workspace.Name,
		"description":     workspace.Description,
		"protection_mode": workspace.ProtectionMode,
		"services":        services,
		"created_at":      workspace.CreatedAt,
		"updated_at":      workspace.UpdatedAt,
	} {
		if err := d.Set(k, v); err != nil {
			return diag.Errorf("error setting %s: %s", k, err)
		}
	}

	return nil
}
