package aclentriescdn

import (
	"context"
	"fmt"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func flattenEntries(ctx context.Context, remoteState []*fastly.ACLEntry, plannedEntries types.Set, diags *diag.Diagnostics) types.Set {
	if len(remoteState) == 0 {
		return types.SetNull(types.ObjectType{AttrTypes: entryAttrTypes})
	}

	// Build two lookup structures:
	//   plannedByKey: full content key → planned entry, for exact matches when entries
	//                 share the same IP but differ in subnet/negated/comment.
	//   plannedByIP:  ip → []planned entries, for fallback null-preservation when
	//                 optional fields are omitted in the plan (null) but the API returns values.
	plannedByKey := make(map[string]EntryModel)
	plannedByIP := make(map[string][]EntryModel)
	if !plannedEntries.IsNull() && !plannedEntries.IsUnknown() {
		var planned []EntryModel
		if d := plannedEntries.ElementsAs(ctx, &planned, false); d.HasError() {
			diags.Append(d...)
		} else {
			for _, e := range planned {
				plannedByKey[plannedEntryContentKey(e)] = e
				if !e.IP.IsNull() && !e.IP.IsUnknown() {
					plannedByIP[e.IP.ValueString()] = append(plannedByIP[e.IP.ValueString()], e)
				}
			}
		}
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

		// Build the content key from the remote entry's actual values.
		ip := ""
		if resource.IP != nil {
			ip = *resource.IP
		}
		subnet := int64(0)
		if resource.Subnet != nil {
			subnet = int64(*resource.Subnet)
		}
		negated := false
		if resource.Negated != nil {
			negated = *resource.Negated
		}
		comment := ""
		if resource.Comment != nil {
			comment = *resource.Comment
		}
		remoteKey := fmt.Sprintf("%s|%d|%t|%s", ip, subnet, negated, comment)

		// Find the matching planned entry.
		// 1. Exact key match handles multiple entries sharing the same IP (e.g. /24 and /32).
		// 2. IP-only fallback handles the case where optional fields are null in the plan
		//    but the API fills them in — valid only when the IP is unambiguous (one entry).
		var plannedEntry *EntryModel
		if pe, ok := plannedByKey[remoteKey]; ok {
			plannedEntry = &pe
		} else if ip != "" {
			if byIP := plannedByIP[ip]; len(byIP) == 1 {
				pe := byIP[0]
				plannedEntry = &pe
			}
		}

		// For optional fields, preserve null from plan if it was null
		if resource.Negated != nil {
			if plannedEntry != nil && plannedEntry.Negated.IsNull() {
				entryAttrs["negated"] = types.BoolNull()
			} else {
				entryAttrs["negated"] = types.BoolValue(*resource.Negated)
			}
		} else if plannedEntry != nil && plannedEntry.Negated.IsNull() {
			entryAttrs["negated"] = types.BoolNull()
		}

		if resource.Comment != nil && *resource.Comment != "" {
			if plannedEntry != nil && plannedEntry.Comment.IsNull() {
				entryAttrs["comment"] = types.StringNull()
			} else {
				entryAttrs["comment"] = types.StringValue(*resource.Comment)
			}
		} else if plannedEntry != nil && plannedEntry.Comment.IsNull() {
			entryAttrs["comment"] = types.StringNull()
		}

		if resource.Subnet != nil {
			if plannedEntry != nil && plannedEntry.Subnet.IsNull() {
				entryAttrs["subnet"] = types.Int64Null()
			} else {
				entryAttrs["subnet"] = types.Int64Value(int64(*resource.Subnet))
			}
		} else if plannedEntry != nil && plannedEntry.Subnet.IsNull() {
			entryAttrs["subnet"] = types.Int64Null()
		}

		obj := types.ObjectValueMust(entryAttrTypes, entryAttrs)
		elements = append(elements, obj)
	}

	set, d := types.SetValue(types.ObjectType{AttrTypes: entryAttrTypes}, elements)
	diags.Append(d...)
	return set
}
