package fastly

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	gofastly "github.com/fastly/go-fastly/v13/fastly"
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

// caCert returns a CA certificate suitable for testing.
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

// generateHex produces a slice of 16 random bytes.
// This is useful for dynamically generating resource names.
func generateHex() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// generateNames produces slice of names seeded with initial unique value.
// e.g. generateNames(generateHex()).
func generateNames(unique string) []string {
	names := []string{}
	for i := 1; i < 4; i++ {
		names = append(names, fmt.Sprintf("tf_%s_%d", unique, i))
	}
	return names
}

// renderTestConfigTemplate is used in acceptance tests to render a
// template and associated data into a Terraform 'configuration file'
// (HCL).
func renderTestConfigTemplate(t *testing.T, tmpl *template.Template, data any) string {
	var output bytes.Buffer

	err := tmpl.Execute(&output, data)
	if err != nil {
		t.Error(t)
	}
	return output.String()
}

// testAccCheckFastlyServiceAttributesBackends is used in acceptance
// tests to compare a list of expected backends against the list of
// configured backends for a service-version.
func testAccCheckFastlyServiceAttributesBackends(service *gofastly.ServiceDetail, name string, backends []string, version int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if gofastly.ToValue(service.Name) != name {
			return fmt.Errorf("bad name, expected (%s), got (%s)", name, gofastly.ToValue(service.Name))
		}

		conn := testAccProvider.Meta().(*APIClient).conn
		backendList, err := conn.ListBackends(context.TODO(), &gofastly.ListBackendsInput{
			ServiceID:      gofastly.ToValue(service.ServiceID),
			ServiceVersion: version,
		})
		if err != nil {
			return fmt.Errorf("error looking up backends for (%s), version (%v): %s", gofastly.ToValue(service.Name), version, err)
		}

		if len(backends) != len(backendList) {
			return fmt.Errorf("backend count mismatch, expected: %#v, got: %#v", len(backends), len(backendList))
		}

		expected := mapset.NewSet[string]()
		expected.Append(backends...)

		found := mapset.NewSet[string]()
		for _, b := range backendList {
			found.Add(gofastly.ToValue(b.Address))
		}

		notExpected := found.Difference(expected)
		notFound := expected.Difference(found)

		var errs []error

		if !notExpected.IsEmpty() {
			errs = append(errs, fmt.Errorf("unexpected backends found: %s", notExpected.String()))
		}

		if !notFound.IsEmpty() {
			errs = append(errs, fmt.Errorf("expected backends not found: %s", notFound.String()))
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	}
}
