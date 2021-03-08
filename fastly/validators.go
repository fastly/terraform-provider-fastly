package fastly

import (
	"encoding/pem"
	"fmt"
	"strings"

	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

// validatePEMBlock returns a schema validation function that checks whether a string contains a single PEM block of
// type `pemType`.
func validatePEMBlock(pemType string) schema.SchemaValidateFunc {
	return func(val interface{}, key string) ([]string, []error) {
		b, rest := pem.Decode([]byte(val.(string)))
		if b == nil {
			return nil, []error{fmt.Errorf("expected %s to be a valid PEM-format block", key)}
		}
		if b.Type != pemType {
			return nil, []error{fmt.Errorf("expected %s to be a valid PEM-format block of type '%s'", key, pemType)}
		}
		if len(rest) != 0 {
			return nil, []error{fmt.Errorf("expected %s to only contain one PEM-format block", key)}
		}
		return nil, nil
	}
}

// validatePEMBlocks returns a schema validation function that checks whether a string contains multiple PEM blocks of
// type `pemType`.
func validatePEMBlocks(pemType string) schema.SchemaValidateFunc {
	return func(val interface{}, key string) ([]string, []error) {
		pemRest := []byte(val.(string))
		numBlocks := 0
		for {
			block, rest := pem.Decode(pemRest)
			if block == nil {
				break
			}
			numBlocks++
			if block.Type != pemType {
				return nil, []error{fmt.Errorf("expected %s to be valid PEM-format blocks of type '%s'", key, pemType)}
			}
			pemRest = rest
		}

		if numBlocks < 1 {
			return nil, []error{fmt.Errorf("expected %s to be valid PEM-format blocks of type '%s'", key, pemType)}
		}

		return nil, nil
	}
}
