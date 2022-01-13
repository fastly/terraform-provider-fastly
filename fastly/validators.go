package fastly

import (
	"encoding/pem"
	"fmt"
	"strings"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validateLoggingFormatVersion() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.IntBetween(1, 2))
}

func validateLoggingMessageType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"classic",
		"loggly",
		"logplex",
		"blank",
	}, false))
}

func validateLoggingCompressionCodec() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"zstd",
		"snappy",
		"gzip",
	}, false))
}

func validateLoggingPlacement() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"none",
		"waf_debug",
	}, false))
}

func validateLoggingServerSideEncryption() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		string(gofastly.S3ServerSideEncryptionAES),
		string(gofastly.S3ServerSideEncryptionKMS),
	}, false))
}

func validateDirectorQuorum() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.IntBetween(0, 100))
}

func validateDirectorType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.IntInSlice([]int{1, 3, 4}))
}

func validateConditionType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"REQUEST",
		"RESPONSE",
		"CACHE",
		"PREFETCH",
	}, false))
}

func validateHeaderAction() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"set",
		"append",
		"delete",
		"regex",
		"regex_repeat",
	}, false))
}

func validateHeaderType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"request",
		"fetch",
		"cache",
		"response",
	}, false))
}

func validateSnippetType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
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
	}, false))
}

func validateRuleStatusType() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice([]string{
		"log",
		"score",
		"block",
	}, false))
}

func validateDictionaryItems() schema.SchemaValidateDiagFunc {
	max := gofastly.MaximumDictionarySize

	return validation.ToDiagFunc(func(i interface{}, k string) (s []string, es []error) {

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
	})
}

func validateUserRole() schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(validation.StringInSlice(
		[]string{
			"user",
			"billing",
			"engineer",
			"superuser",
		},
		false,
	))
}

// validatePEMBlock returns a schema validation function that checks whether a string contains a single PEM block of
// type `pemType`.
func validatePEMBlock(pemType string) schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(func(val interface{}, key string) ([]string, []error) {
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
	})
}

// validatePEMBlocks returns a schema validation function that checks whether a string contains multiple PEM blocks of
// type `pemType`.
func validatePEMBlocks(pemType string) schema.SchemaValidateDiagFunc {
	return validation.ToDiagFunc(func(val interface{}, key string) ([]string, []error) {
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
	})
}

func validateStringTrimmed(i interface{}, path cty.Path) diag.Diagnostics {
	v := i.(string)
	attr := path[len(path)-1].(cty.GetAttrStep)
	if v != strings.TrimSpace(v) {
		return diag.Errorf("%s must not contain trailing space characters (e.g., \\n\\t\\r\\f). Consider using trimspace() function", attr.Name)
	}

	return nil
}
