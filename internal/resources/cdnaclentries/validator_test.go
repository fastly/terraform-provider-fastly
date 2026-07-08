package cdnaclentries

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func entryObj(id, ip string, subnet int64, negated bool, comment string) attr.Value {
	return types.ObjectValueMust(entryAttrTypes, map[string]attr.Value{
		"id": types.StringValue(id), "ip": types.StringValue(ip),
		"subnet": types.Int64Value(subnet), "negated": types.BoolValue(negated), "comment": types.StringValue(comment),
	})
}

func TestUniqueEntryIdentity_RejectsDuplicateIPAndSubnet(t *testing.T) {
	ctx := context.Background()

	set := types.SetValueMust(types.ObjectType{AttrTypes: entryAttrTypes}, []attr.Value{
		entryObj("", "10.0.0.1", 24, false, "first"),
		entryObj("", "10.0.0.1", 24, true, "second, differs only by negated/comment"),
	})

	req := validator.SetRequest{Path: path.Root("entry"), ConfigValue: set}
	resp := &validator.SetResponse{}

	UniqueEntryIdentity().ValidateSet(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUniqueEntryIdentity_AllowsDifferentSubnetSameIP(t *testing.T) {
	ctx := context.Background()

	set := types.SetValueMust(types.ObjectType{AttrTypes: entryAttrTypes}, []attr.Value{
		entryObj("", "10.0.0.1", 24, false, "a"),
		entryObj("", "10.0.0.1", 32, false, "b"),
	})

	req := validator.SetRequest{Path: path.Root("entry"), ConfigValue: set}
	resp := &validator.SetResponse{}

	UniqueEntryIdentity().ValidateSet(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUniqueEntryIdentity_NullAndUnknownAreNoOp(t *testing.T) {
	ctx := context.Background()

	for _, set := range []types.Set{
		types.SetNull(types.ObjectType{AttrTypes: entryAttrTypes}),
		types.SetUnknown(types.ObjectType{AttrTypes: entryAttrTypes}),
	} {
		req := validator.SetRequest{Path: path.Root("entry"), ConfigValue: set}
		resp := &validator.SetResponse{}

		UniqueEntryIdentity().ValidateSet(ctx, req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	}
}
