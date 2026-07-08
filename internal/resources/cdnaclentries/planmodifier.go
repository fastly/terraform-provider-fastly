package cdnaclentries

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type preserveEntryIDsModifier struct{}

func (m preserveEntryIDsModifier) Description(_ context.Context) string {
	return "Preserves entry IDs from state when entries match by content."
}

func (m preserveEntryIDsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m preserveEntryIDsModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		return
	}

	var stateEntries []EntryModel
	var planEntries []EntryModel

	diags := req.StateValue.ElementsAs(ctx, &stateEntries, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.PlanValue.ElementsAs(ctx, &planEntries, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateByContent := make(map[string]string)
	for _, entry := range stateEntries {
		if entry.ID.IsNull() || entry.ID.IsUnknown() {
			continue
		}
		key := entryContentKey(entry)
		stateByContent[key] = entry.ID.ValueString()
	}

	modified := false
	for i := range planEntries {
		if !planEntries[i].ID.IsUnknown() {
			continue
		}

		key := entryContentKey(planEntries[i])
		if existingID, found := stateByContent[key]; found {
			planEntries[i].ID = types.StringValue(existingID)
			modified = true
		}
	}

	if !modified {
		return
	}

	updatedSet, diags := types.SetValueFrom(ctx, req.PlanValue.ElementType(ctx), planEntries)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = updatedSet
}

func entryContentKey(e EntryModel) string {
	ip := ""
	subnet := int64(0)
	negated := false
	comment := ""

	if !e.IP.IsNull() && !e.IP.IsUnknown() {
		ip = e.IP.ValueString()
	}
	if !e.Subnet.IsNull() && !e.Subnet.IsUnknown() {
		subnet = e.Subnet.ValueInt64()
	}
	if !e.Negated.IsNull() && !e.Negated.IsUnknown() {
		negated = e.Negated.ValueBool()
	}
	if !e.Comment.IsNull() && !e.Comment.IsUnknown() {
		comment = e.Comment.ValueString()
	}

	return fmt.Sprintf("%s|%d|%t|%s", ip, subnet, negated, comment)
}
