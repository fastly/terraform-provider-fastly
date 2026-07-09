package aclentries

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	ID            types.String `tfsdk:"id"`
	ACLID         types.String `tfsdk:"acl_id"`
	Entries       types.Map    `tfsdk:"entries"`
	ManageEntries types.Bool   `tfsdk:"manage_entries"`
}

func ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "Terraform resource identifier. Format: `acl_id/entries`.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"acl_id": schema.StringAttribute{
			Required:    true,
			Description: "The ID of the ACL that the entries belong to.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"entries": schema.MapAttribute{
			Required:    true,
			ElementType: types.StringType,
			Description: "A map representing the entries in the ACL, where the keys are CIDR prefixes and the values are actions (`ALLOW` or `BLOCK`).",
			Validators: []validator.Map{
				ValidEntries(),
			},
		},
		"manage_entries": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Manage the ACL entries in Terraform (default: `false`). If `true`, Terraform will ensure that the ACL's entries match the entries in the Terraform configuration.",
		},
	}
}
