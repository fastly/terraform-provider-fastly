package resourcelink

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

var linkAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"resource_id": types.StringType,
	"link_id":     types.StringType,
}

func linkObj(name, resourceID, linkID string) attr.Value {
	return types.ObjectValueMust(linkAttrTypes, map[string]attr.Value{
		"name": types.StringValue(name), "resource_id": types.StringValue(resourceID), "link_id": types.StringValue(linkID),
	})
}

func TestUniqueResourceLinkIdentity_RejectsDuplicateName(t *testing.T) {
	ctx := context.Background()

	list := types.ListValueMust(types.ObjectType{AttrTypes: linkAttrTypes}, []attr.Value{
		linkObj("store", "resource_a", ""),
		linkObj("store", "resource_b", ""),
	})

	req := validator.ListRequest{Path: path.Root("resource_link"), ConfigValue: list}
	resp := &validator.ListResponse{}

	UniqueResourceLinkIdentity().ValidateList(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUniqueResourceLinkIdentity_RejectsDuplicateResourceID(t *testing.T) {
	ctx := context.Background()

	list := types.ListValueMust(types.ObjectType{AttrTypes: linkAttrTypes}, []attr.Value{
		linkObj("alias_a", "resource_1", ""),
		linkObj("alias_b", "resource_1", ""),
	})

	req := validator.ListRequest{Path: path.Root("resource_link"), ConfigValue: list}
	resp := &validator.ListResponse{}

	UniqueResourceLinkIdentity().ValidateList(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
}

func TestUniqueResourceLinkIdentity_AllowsDistinctNameAndResourceID(t *testing.T) {
	ctx := context.Background()

	list := types.ListValueMust(types.ObjectType{AttrTypes: linkAttrTypes}, []attr.Value{
		linkObj("alias_a", "resource_1", ""),
		linkObj("alias_b", "resource_2", ""),
	})

	req := validator.ListRequest{Path: path.Root("resource_link"), ConfigValue: list}
	resp := &validator.ListResponse{}

	UniqueResourceLinkIdentity().ValidateList(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestUniqueResourceLinkIdentity_NullAndUnknownAreNoOp(t *testing.T) {
	ctx := context.Background()

	for _, list := range []types.List{
		types.ListNull(types.ObjectType{AttrTypes: linkAttrTypes}),
		types.ListUnknown(types.ObjectType{AttrTypes: linkAttrTypes}),
	} {
		req := validator.ListRequest{Path: path.Root("resource_link"), ConfigValue: list}
		resp := &validator.ListResponse{}

		UniqueResourceLinkIdentity().ValidateList(ctx, req, resp)

		assert.False(t, resp.Diagnostics.HasError())
	}
}
