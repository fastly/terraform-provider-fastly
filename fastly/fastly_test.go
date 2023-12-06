package fastly

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_publickey", t)
}

// privatekey returns a ASN.1 DER encoded key suitable for testing.
func privateKey(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_privatekey", t)
}

// certificate returns a ASN.1 DER encoded certificate for the private key suitable for testing.
func certificate(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_certificate", t)
}

// caCert returns a CA certificate suitable for testing
func caCert(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_cacert", t)
}

func readTestFile(filename string, t *testing.T) string {
	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Cannot load key file: %s", filename)
	}
	return string(contents)
}

// escapePercentSign uses Terraform's escape syntax (i.e., repeating characters)
// to properly escape percent signs (i.e., '%').
//
// There are two significant places where '%' can show up:
// 1. Before a left curly brace (i.e., '{').
// 2. Not before a left curly brace.
//
// In case #1, we have to double escape so that Terraform does not interpret Fastly's
// configuration values as its own (e.g., https://docs.fastly.com/en/guides/custom-log-formats).
//
// In case #2, we only have to single escape.
//
// Refer to the Terraform documentation on string literals for more details:
// https://www.terraform.io/docs/configuration/expressions.html#string-literals
func escapePercentSign(s string) string {
	escapeSign := strings.ReplaceAll(s, "%", "%%")
	escapeTemplateSequence := strings.ReplaceAll(escapeSign, "%%{", "%%%%{")

	return escapeTemplateSequence
}

func TestEscapePercentSign(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "string no percent signs should change nothing",
			input: "abc 123",
			want:  "abc 123",
		},
		{
			name:  "one percent sign should return two percent signs",
			input: "%",
			want:  "%%",
		},
		{
			name:  "one percent sign mid-string should return two percent signs in the same place",
			input: "abc%123",
			want:  "abc%%123",
		},
		{
			name:  "one percent sign before left curly brace should return four percent signs then curly brace",
			input: "%{",
			want:  "%%%%{",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got := escapePercentSign(testcase.input)

			if got != testcase.want {
				t.Errorf("escapePercentSign(%q): \n\tgot: '%+v'\n\twant: '%+v'", testcase.input, got, testcase.want)
			}
		})
	}
}

func appendNewLine(s string) string {
	return s + "\n"
}

// assertEqualsSliceOfMaps compares a slice of maps even if they include schema.Set values
func assertEqualsSliceOfMaps(t *testing.T, actualSlice []map[string]any, expectedSlice []map[string]any) {
	for i, actualMap := range actualSlice {
		var keysToBeRemoved []string
		for key, value := range actualMap {
			if v, ok := value.(*schema.Set); ok {
				expected := expectedSlice[i][key]
				keysToBeRemoved = append(keysToBeRemoved, key)
				if !v.Equal(expected) {
					t.Errorf("expected sets %s to be equal: %#v\n     got: %#v", key, expected, actualSlice)
				}
			}
		}
		for _, key := range keysToBeRemoved {
			delete(actualMap, key)
			delete(expectedSlice[i], key)
		}
	}

	if !reflect.DeepEqual(actualSlice, expectedSlice) {
		t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", expectedSlice, actualSlice)
	}
}

// generateHex produces a slice of 16 random bytes.
// This is useful for dynamically generating resource names.
func generateHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// generateNames produces slice of names seeded with initial unique value.
// e.g. generateNames(generateHex(), 3)
func generateNames(unique string, size int) []string {
	names := []string{}
	for i := 1; i < size+1; i++ {
		names = append(names, fmt.Sprintf("tf_%s_%d", unique, i))
	}
	return names
}
