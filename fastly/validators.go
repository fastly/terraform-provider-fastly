package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

func validateLoggingServerSideEncryption() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		gofastly.S3ServerSideEncryptionAES,
		gofastly.S3ServerSideEncryptionKMS,
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
		"PREFETCH",
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

func validateDictionaryItems() schema.SchemaValidateFunc {

	max := gofastly.MaximumDictionarySize

	return func(i interface{}, k string) (s []string, es []error) {

		v, ok := i.(map[string]interface{})
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be a map[string]interface", k))
			return
		}

		if len(v) > max {
			es = append(es, fmt.Errorf("expected %s to be at most (%d), got %d", k, max, len(v)))
			return
		}

		return
	}

}
