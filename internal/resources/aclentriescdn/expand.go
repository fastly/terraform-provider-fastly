package aclentriescdn

import (
	"context"
	"fmt"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func buildBatchACLEntry(ctx context.Context, entry EntryModel, op fastly.BatchOperation) *fastly.BatchACLEntry {
	batchEntry := &fastly.BatchACLEntry{
		Operation: &op,
		IP:        entry.IP.ValueStringPointer(),
		Comment:   entry.Comment.ValueStringPointer(),
	}

	if !entry.ID.IsNull() && !entry.ID.IsUnknown() {
		batchEntry.EntryID = entry.ID.ValueStringPointer()
	}

	if !entry.Negated.IsNull() && !entry.Negated.IsUnknown() {
		negatedValue := entry.Negated.ValueBool()
		negated := fastly.Compatibool(negatedValue)
		batchEntry.Negated = &negated
		tflog.Debug(ctx, "Building batch ACL entry", map[string]any{
			"ip":             entry.IP.ValueString(),
			"negated":        negatedValue,
			"negated_compat": fmt.Sprintf("%v", negated),
			"op":             op,
		})
	} else {
		tflog.Debug(ctx, "Building batch ACL entry", map[string]any{
			"ip":      entry.IP.ValueString(),
			"negated": "not set (will use API default)",
			"op":      op,
		})
	}

	if !entry.Subnet.IsNull() && !entry.Subnet.IsUnknown() {
		subnet := int(entry.Subnet.ValueInt64())
		batchEntry.Subnet = &subnet
	}

	return batchEntry
}

func expandEntries(ctx context.Context, entries types.Set, diags *diag.Diagnostics) []EntryModel {
	if entries.IsNull() || entries.IsUnknown() {
		return nil
	}

	var entryModels []EntryModel
	diags.Append(entries.ElementsAs(ctx, &entryModels, false)...)
	return entryModels
}

func entryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"ip":      types.StringType,
		"subnet":  types.Int64Type,
		"negated": types.BoolType,
		"comment": types.StringType,
	}
}
