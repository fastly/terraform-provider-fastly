package productenablement

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// idAttribute and serviceIDAttribute are shared by every product-enablement
// resource: each one is versionless (the underlying Fastly Product
// Enablement APIs are scoped only by service_id, with no concept of a
// service version) and keyed on service_id alone.
func idAttribute() schema.Attribute {
	return schema.StringAttribute{
		Computed:    true,
		Description: "The ID of this resource (identical to `service_id`).",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func serviceIDAttribute(displayName string) schema.Attribute {
	return schema.StringAttribute{
		Required:    true,
		Description: fmt.Sprintf("The ID of the service to enable %s on. Changing this value will delete and recreate this resource against the new service.", displayName),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// isEntitlementError reports whether err is the Fastly API's way of saying
// the caller isn't entitled to disable a product (either because they never
// had self-service access, or because the product forbids disabling
// entirely). These are treated as success so that a `terraform destroy`
// can still complete and leave a clean state, even for accounts where
// Fastly support manages product enablement manually.
func isEntitlementError(err error) bool {
	var httpErr *fastly.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		return false
	}
	for _, e := range httpErr.Errors {
		if strings.Contains(e.Title, "not entitled to disable") ||
			strings.Contains(e.Title, "product cannot be disabled") ||
			strings.Contains(e.Title, "cannot self-disable") {
			return true
		}
	}
	return false
}
