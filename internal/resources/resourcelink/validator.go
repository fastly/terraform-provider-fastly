package resourcelink

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// uniqueResourceLinkIdentityValidator rejects config where two or more resource_link
// blocks share the same name or the same resource_id. Reconciliation keys resource
// links by resource_id, so a duplicate there would make blocks indistinguishable at
// apply time; a duplicate name would let two links use the same Compute-code alias.
type uniqueResourceLinkIdentityValidator struct{}

func UniqueResourceLinkIdentity() validator.List {
	return uniqueResourceLinkIdentityValidator{}
}

func (v uniqueResourceLinkIdentityValidator) Description(_ context.Context) string {
	return "Resource links must have unique name and resource_id values."
}

func (v uniqueResourceLinkIdentityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueResourceLinkIdentityValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var links []NestedModel
	diags := req.ConfigValue.ElementsAs(ctx, &links, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	namesSeenAt := make(map[string]int)
	resourceIDsSeenAt := make(map[string]int)
	for i, link := range links {
		if link.Name.IsUnknown() || link.ResourceID.IsUnknown() {
			continue
		}

		name := link.Name.ValueString()
		if first, ok := namesSeenAt[name]; ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate Resource Link Name",
				fmt.Sprintf(
					"Resource links at index %d and %d both use the name %q. Each link's name is the alias Compute code uses to open it, so it must be unique.",
					first, i, name,
				),
			)
		} else {
			namesSeenAt[name] = i
		}

		resourceID := link.ResourceID.ValueString()
		if first, ok := resourceIDsSeenAt[resourceID]; ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate Resource Link",
				fmt.Sprintf(
					"Resource links at index %d and %d both target resource_id %q. Each shared resource can only be linked once per service version.",
					first, i, resourceID,
				),
			)
		} else {
			resourceIDsSeenAt[resourceID] = i
		}
	}
}
