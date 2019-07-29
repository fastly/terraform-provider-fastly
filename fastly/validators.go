package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	gofastly "github.com/fastly/go-fastly/fastly"
)

func validateLoggingFormatVersion() schema.SchemaValidateFunc {
	return validation.IntBetween(1, 2)
}

func validateLoggingMessageType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"classic",
		"loggly",
		"logplex",
		"blank",
	}, false)
}

func validateLoggingPlacement() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"none",
		"waf_debug",
	}, false)
}

func validateDirectorQuorum() schema.SchemaValidateFunc {
	return validation.IntBetween(0, 100)
}

func validateDirectorType() schema.SchemaValidateFunc {
	return validation.IntInSlice([]int{1, 3, 4})
}

func validateConditionType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"REQUEST",
		"RESPONSE",
		"CACHE",
	}, false)
}

func validateHeaderAction() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"set",
		"append",
		"delete",
		"regex",
		"regex_repeat",
	}, false)
}

func validateHeaderType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"request",
		"fetch",
		"cache",
		"response",
	}, false)
}

func validateSnippetType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"init",
		"recv",
		"hit",
		"miss",
		"pass",
		"fetch",
		"error",
		"deliver",
		"log",
		"none",
	}, false)
}

func validateACLEntries() schema.SchemaValidateFunc {

	max := gofastly.MaximumACLSize

	return func(i interface{}, k string) (s []string, es []error) {

		v, ok := i.(*schema.Set)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be a schema.Set", k))
			return
		}

		if len(v.List()) > max {
			es = append(es, fmt.Errorf("expected %s to be at most (%d), got %d", k, max, len(v.List())))
			return
		}

		return
	}

}
