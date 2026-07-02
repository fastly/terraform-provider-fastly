package aclentriescdn

import (
	"context"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func flattenEntries(ctx context.Context, remoteState []*fastly.ACLEntry, diags *diag.Diagnostics) types.Set {
	if len(remoteState) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: entryAttrTypes()})
	}

	var elements []attr.Value

	for _, resource := range remoteState {
		entryAttrs := map[string]attr.Value{
			"id":      types.StringNull(),
			"ip":      types.StringNull(),
			"subnet":  types.Int64Null(),
			"negated": types.BoolValue(false),
			"comment": types.StringNull(),
		}

		if resource.EntryID != nil {
			entryAttrs["id"] = types.StringValue(*resource.EntryID)
		}
		if resource.IP != nil {
			entryAttrs["ip"] = types.StringValue(*resource.IP)
		}
		if resource.Negated != nil {
			entryAttrs["negated"] = types.BoolValue(*resource.Negated)
			tflog.Debug(ctx, "Flatten: entry negated value", map[string]any{
				"ip":            resource.IP,
				"negated_raw":   *resource.Negated,
				"negated_typed": types.BoolValue(*resource.Negated),
			})
		}
		if resource.Comment != nil && *resource.Comment != "" {
			entryAttrs["comment"] = types.StringValue(*resource.Comment)
		}
		if resource.Subnet != nil {
			entryAttrs["subnet"] = types.Int64Value(int64(*resource.Subnet))
		}

		obj := types.ObjectValueMust(entryAttrTypes(), entryAttrs)
		elements = append(elements, obj)
	}

	set, d := types.SetValue(types.ObjectType{AttrTypes: entryAttrTypes()}, elements)
	diags.Append(d...)
	return set
}
