package cdnaclentries

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// uniqueEntryIdentityValidator rejects config where two or more entry blocks
// share the same (ip, subnet) -- the identity Fastly's ACL enforces
// uniqueness on. Without this, two blocks that differ only in comment or
// negated would pass schema validation but collide at apply time, since
// Update() treats them as the same entry.
type uniqueEntryIdentityValidator struct{}

func UniqueEntryIdentity() validator.Set {
	return uniqueEntryIdentityValidator{}
}

func (v uniqueEntryIdentityValidator) Description(_ context.Context) string {
	return "Entries must have unique ip + subnet combinations."
}

func (v uniqueEntryIdentityValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v uniqueEntryIdentityValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var entries []EntryModel
	diags := req.ConfigValue.ElementsAs(ctx, &entries, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	seenAt := make(map[string]int)
	for i, entry := range entries {
		if entry.IP.IsUnknown() || entry.Subnet.IsUnknown() {
			continue
		}

		key := entryIdentityKey(entry)
		if first, ok := seenAt[key]; ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate ACL Entry",
				fmt.Sprintf(
					"Entries at index %d and %d both target IP %q with subnet %d. Fastly's ACL enforces uniqueness on ip + subnet -- use a single entry to change its comment or negated field.",
					first, i, entry.IP.ValueString(), entry.Subnet.ValueInt64(),
				),
			)
			continue
		}
		seenAt[key] = i
	}
}
