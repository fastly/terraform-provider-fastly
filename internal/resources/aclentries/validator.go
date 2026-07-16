package aclentries

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// validEntriesValidator checks that every key in the entries map is a valid
// CIDR prefix and every value is one of the actions the ACL API
// accepts.
type validEntriesValidator struct{}

func ValidEntries() validator.Map {
	return validEntriesValidator{}
}

func (v validEntriesValidator) Description(_ context.Context) string {
	return "Keys must be valid CIDR prefixes and values must be either ALLOW or BLOCK."
}

func (v validEntriesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v validEntriesValidator) ValidateMap(_ context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for prefix, value := range req.ConfigValue.Elements() {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid ACL Entry Prefix",
				fmt.Sprintf("%q is not a valid CIDR prefix: %s", prefix, err),
			)
		}

		strVal, ok := value.(types.String)
		if !ok || strVal.IsUnknown() || strVal.IsNull() {
			continue
		}

		if action := strVal.ValueString(); action != "ALLOW" && action != "BLOCK" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid ACL Entry Action",
				fmt.Sprintf("action %q for prefix %q must be either ALLOW or BLOCK", action, prefix),
			)
		}
	}
}
