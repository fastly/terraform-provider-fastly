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
	validPlacements := map[string]struct{}{
		"none":      {},
		"waf_debug": {},
	}

	if _, ok := validPlacements[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['none', 'waf_debug']", k))
	}
	return
}

func validateDirectorQuorum(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 100 {
		errors = append(errors, fmt.Errorf(
			"%q must be a percentage between 0 and 100", k))
	}
	return
}

func validateDirectorType(v interface{}, k string) (ws []string, errors []error) {
	value := uint(v.(int))
	validTypes := map[uint]struct{}{
		1: {},
		3: {},
		4: {},
	}

	if _, ok := validTypes[value]; !ok {
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

func validateHeaderAction(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validActions := map[string]struct{}{
		"set":          {},
		"append":       {},
		"delete":       {},
		"regex":        {},
		"regex_repeat": {},
	}

	if _, ok := validActions[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['set', 'append', 'delete', 'regex', 'regex_repeat']", k))
	}
	return
}

func validateHeaderType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"request":  {},
		"fetch":    {},
		"cache":    {},
		"response": {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['request', 'fetch', 'cache', 'response']", k))
	}
	return
}

func validateSnippetType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"init":    {},
		"recv":    {},
		"hit":     {},
		"miss":    {},
		"pass":    {},
		"fetch":   {},
		"error":   {},
		"deliver": {},
		"log":     {},
		"none":    {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['init', 'recv', 'hit', 'miss', 'pass', 'fetch', 'error', 'deliver', 'log', 'none']", k))
	}
	return
}

func validateHeaderAction(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validActions := map[string]struct{}{
		"set":          {},
		"append":       {},
		"delete":       {},
		"regex":        {},
		"regex_repeat": {},
	}

	if _, ok := validActions[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['set', 'append', 'delete', 'regex', 'regex_repeat']", k))
	}
	return
}

func validateHeaderType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"request":  {},
		"fetch":    {},
		"cache":    {},
		"response": {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['request', 'fetch', 'cache', 'response']", k))
	}
	return
}

func validateSnippetType(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validTypes := map[string]struct{}{
		"init":    {},
		"recv":    {},
		"hit":     {},
		"miss":    {},
		"pass":    {},
		"fetch":   {},
		"error":   {},
		"deliver": {},
		"log":     {},
		"none":    {},
	}

	if _, ok := validTypes[value]; !ok {
		errors = append(errors, fmt.Errorf(
			"%q must be one of ['init', 'recv', 'hit', 'miss', 'pass', 'fetch', 'error', 'deliver', 'log', 'none']", k))
	}
	return
}
