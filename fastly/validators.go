package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func validateLoggingFormatVersion(v interface{}, k string) (ws []string, errors []error) {
	value := uint(v.(int))
	validVersions := map[uint]struct{}{
		1: {},
		2: {},
	}

	if _, ok := validVersions[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['1', '2']", k))
	}
	return
}

func validateLoggingMessageType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"classic": {},
		"loggly":  {},
		"logplex": {},
		"blank":   {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['classic', 'loggly', 'logplex', 'blank']", k))
	}
	return
}

func validateLoggingPlacement(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"none":      {},
		"waf_debug": {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['none', 'waf_debug']", k))
	}
	return
}

func validateDirectorQuorum() schema.SchemaValidateFunc {
	return validation.IntBetween(0, 100)
}

func validateDirectorType(v interface{}, k string) (ws []string, errors []error) {
	value := uint(v.(int))
	validVersions := map[uint]struct{}{
		1: {},
		3: {},
		4: {},
	}

	if _, ok := validVersions[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['1' (random), '3' (hash), '4' (client)]", k))
	}
	return
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
