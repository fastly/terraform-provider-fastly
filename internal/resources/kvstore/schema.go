package kvstore

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:    true,
			Description: "The ID of this KV Store.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "A unique name to identify the KV Store. Changing this attribute will delete and recreate the KV Store, discarding its current entries. Any `resource_link` referencing this KV Store must be removed first.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"location": schema.StringAttribute{
			Optional:    true,
			Description: "The regional location of the KV Store. Valid values are `US`, `EU`, `ASIA`, and `AUS`. Changing this attribute will delete and recreate the KV Store.",
			Validators: []validator.String{
				stringvalidator.OneOf("US", "EU", "ASIA", "AUS"),
			},
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"force_destroy": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
			Description: "Allow the KV Store to be deleted, even if it contains entries. Defaults to false.",
		},
	}
}
