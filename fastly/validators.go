package fastly

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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

func validateBackendWeight() schema.SchemaValidateFunc {
	return validation.IntBetween(1, 100)
}
