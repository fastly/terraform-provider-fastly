package computepackage

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name        string
		packages    []Model
		expectError bool
	}{
		{
			name:        "empty packages",
			packages:    []Model{},
			expectError: false,
		},
		{
			name: "multiple packages",
			packages: []Model{
				{Filename: types.StringValue("pkg1.tar.gz")},
				{Filename: types.StringValue("pkg2.tar.gz")},
			},
			expectError: true,
		},
		{
			name: "both content and filename set",
			packages: []Model{{
				Content:  types.StringValue("content"),
				Filename: types.StringValue("file.tar.gz"),
			}},
			expectError: true,
		},
		{
			name: "neither content nor filename set",
			packages: []Model{{
				Content:  types.StringNull(),
				Filename: types.StringNull(),
			}},
			expectError: true,
		},
		{
			name: "valid filename",
			packages: []Model{{
				Filename: types.StringValue("package.tar.gz"),
			}},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packages := tt.packages
			err := ValidateInput(packages)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []Model
		b        []Model
		expected bool
	}{
		{
			name:     "both empty",
			a:        []Model{},
			b:        []Model{},
			expected: true,
		},
		{
			name:     "different lengths",
			a:        []Model{{}},
			b:        []Model{},
			expected: false,
		},
		{
			name: "same content",
			a: []Model{{
				Content: types.StringValue("content1"),
			}},
			b: []Model{{
				Content: types.StringValue("content1"),
			}},
			expected: true,
		},
		{
			name: "same filename",
			a: []Model{{
				Filename: types.StringValue("package.tar.gz"),
			}},
			b: []Model{{
				Filename: types.StringValue("package.tar.gz"),
			}},
			expected: true,
		},
		{
			name: "different content",
			a: []Model{{
				Content: types.StringValue("content1"),
			}},
			b: []Model{{
				Content: types.StringValue("content2"),
			}},
			expected: false,
		},
		{
			name: "different filename",
			a: []Model{{
				Filename: types.StringValue("package1.tar.gz"),
			}},
			b: []Model{{
				Filename: types.StringValue("package2.tar.gz"),
			}},
			expected: false,
		},
		{
			name: "same content and filename, no source_code_hash set",
			a: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringNull(),
			}},
			b: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringNull(),
			}},
			expected: true,
		},
		{
			name: "same content and filename, different source_code_hash when set",
			a: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringValue("hash1"),
			}},
			b: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringValue("hash2"),
			}},
			expected: false, // Should detect difference when hash is explicitly set
		},
		{
			name: "same content and filename, same source_code_hash when set",
			a: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringValue("hash1"),
			}},
			b: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringValue("hash1"),
			}},
			expected: true,
		},
		{
			name: "same content and filename, one has hash set and one doesn't",
			a: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringValue("hash1"),
			}},
			b: []Model{{
				Content:        types.StringValue("content1"),
				Filename:       types.StringValue("package.tar.gz"),
				SourceCodeHash: types.StringNull(),
			}},
			expected: true, // Should ignore hash when only one side has it (API-populated)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equal(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Equal() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestStringValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        types.String
		b        types.String
		expected bool
	}{
		{
			name:     "both null",
			a:        types.StringNull(),
			b:        types.StringNull(),
			expected: true,
		},
		{
			name:     "both unknown",
			a:        types.StringUnknown(),
			b:        types.StringUnknown(),
			expected: true,
		},
		{
			name:     "same values",
			a:        types.StringValue("test"),
			b:        types.StringValue("test"),
			expected: true,
		},
		{
			name:     "different values",
			a:        types.StringValue("test1"),
			b:        types.StringValue("test2"),
			expected: false,
		},
		{
			name:     "one null one value",
			a:        types.StringNull(),
			b:        types.StringValue("test"),
			expected: false,
		},
		{
			name:     "one unknown one value",
			a:        types.StringUnknown(),
			b:        types.StringValue("test"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringValuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("stringValuesEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
