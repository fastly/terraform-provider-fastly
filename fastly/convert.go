package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func uintOrDefault(int *uint) uint {
	if int == nil {
		return 0
	}
	return *int
}

// diagToErr takes a diag.Diagnostics and finds the first Error (ignoring Warnings).
// This is useful for some of the SDK functions which are context aware but still return Go errors, e.g. StateContext
// and resource.RetryContext.
func diagToErr(diagnostics diag.Diagnostics) error {
	if diagnostics.HasError() {
		// diagnostics could have multiple Warnings as well as an Error
		for _, diagnostic := range diagnostics {
			if diagnostic.Severity == diag.Error {
				return fmt.Errorf("%s", diagnostic.Summary)
			}
		}
	}
	return nil
}
