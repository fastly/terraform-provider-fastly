package fastly

import (
	"encoding/pem"
	"fmt"
	"strings"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
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
		string(gofastly.S3ServerSideEncryptionAES),
		string(gofastly.S3ServerSideEncryptionKMS),
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
		"hash",
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

func validateRuleStatusType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"log",
		"score",
		"block",
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

func validateUserRole() schema.SchemaValidateFunc {
	return validation.StringInSlice(
		[]string{
			"user",
			"billing",
			"engineer",
			"superuser",
		},
		false,
	)
}

// TODO: Use SDK's validation.IsURLWithHTTPS() after we upgrade
func validateHTTPSURL() schema.SchemaValidateFunc {
	return func(val interface{}, key string) (warns []string, errs []error) {
		v := val.(string)
		if !strings.HasPrefix(v, "https://") {
			errs = append(errs, fmt.Errorf("%q must be https URL, got: %s", key, v))
		}
		return
	}
}

func validatePEMBlock(pemType string) schema.SchemaValidateFunc {
	return func(val interface{}, key string) ([]string, []error) {
		b, _ := pem.Decode([]byte(val.(string))) // TODO: might want to loop here to support multiple PEM resources in one string for intermediates blob?
		if b == nil {
			return nil, []error{fmt.Errorf("expected %s to be a valid PEM-format block", key)}
		}
		if b.Type != pemType {
			return nil, []error{fmt.Errorf("expected %s to be a valid PEM-format block of type '%s'", key, pemType)}
		}
		return nil, nil
	}
}
