package dynamicsnippetcontent

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	ID            types.String `tfsdk:"id"`
	Service       types.String `tfsdk:"service_id"`
	SnippetID     types.String `tfsdk:"snippet_id"`
	Content       types.String `tfsdk:"content"`
	ManageSnippet types.Bool   `tfsdk:"manage_snippets"`
}

func ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier.",
		},
		"service_id": schema.StringAttribute{
			Required:    true,
			Description: "Fastly service ID.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"snippet_id": schema.StringAttribute{
			Required:    true,
			Description: "The Fastly-generated ID of the dynamic VCL snippet whose content is managed by this resource.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"content": schema.StringAttribute{
			Required:    true,
			Description: "The dynamic VCL snippet code. Updates are versionless and take effect immediately without cloning or activating a service version.",
		},
		"manage_snippets": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Whether Terraform should re-apply content drift and clear snippet content on destroy. Default `false`.",
		},
	}
}

func ID(serviceID string, snippetID string) string {
	return serviceID + "/" + snippetID
}
