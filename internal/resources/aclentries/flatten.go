package aclentries

import (
	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func flattenEntries(remote []computeacls.ComputeACLEntry, diags *diag.Diagnostics) types.Map {
	elements := make(map[string]attr.Value, len(remote))
	for _, e := range remote {
		elements[e.Prefix] = types.StringValue(e.Action)
	}

	m, d := types.MapValue(types.StringType, elements)
	diags.Append(d...)
	return m
}
