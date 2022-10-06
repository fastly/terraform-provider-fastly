package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestSetDiffDiff(t *testing.T) {
	cases := []struct {
		name               string
		keyFunc            KeyFunc
		oldElements        []map[string]any
		newElements        []map[string]any
		expectedAdded      []map[string]any
		expectedModified   []map[string]any
		expectedDeleted    []map[string]any
		expectedUnmodified []map[string]any
		expectedError      bool
	}{
		{
			name: "should return the correct diff",
			oldElements: []map[string]any{
				{
					"name":  "name-a",
					"value": "value-a",
				},
				{
					"name":  "b",
					"value": "value-b",
				},
				{
					"name":  "name-d",
					"value": "value-d",
				},
			},
			newElements: []map[string]any{
				{
					"name":  "name-a",
					"value": "value-a-new",
				},
				{
					"name":  "name-c",
					"value": "value-c",
				},
				{
					"name":  "name-d",
					"value": "value-d",
				},
			},
			expectedAdded: []map[string]any{
				{
					"name":  "name-c",
					"value": "value-c",
				},
			},
			expectedModified: []map[string]any{
				{
					"name":  "name-a",
					"value": "value-a-new",
				},
			},
			expectedDeleted: []map[string]any{
				{
					"name":  "b",
					"value": "value-b",
				},
			},
			expectedUnmodified: []map[string]any{
				{
					"name":  "name-d",
					"value": "value-d",
				},
			},
		},
		{
			name: "should diff empty element lists",
		},
		{
			name: "should return error if key cannot be computed",
			oldElements: []map[string]any{
				{},
			},
			newElements: []map[string]any{
				{},
			},
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf(c.name), func(t *testing.T) {
			var differ *SetDiff
			if c.keyFunc != nil {
				differ = NewSetDiff(c.keyFunc)
			} else {
				differ = NewSetDiff(testKeyFuncByName)
			}

			diff, err := differ.Diff(testCreateSet(c.oldElements), testCreateSet(c.newElements))

			if err != nil && !c.expectedError {
				t.Fatalf("Error not expected: %v", err)
			}

			if err == nil && c.expectedError {
				t.Fatalf("Error expected: %v", err)
			}

			if err == nil && !c.expectedError {
				assert.ElementsMatch(t, c.expectedAdded, diff.Added)
				assert.ElementsMatch(t, c.expectedModified, diff.Modified)
				assert.ElementsMatch(t, c.expectedDeleted, diff.Deleted)
				assert.ElementsMatch(t, c.expectedUnmodified, diff.Unmodified)
			}
		})
	}
}

func testKeyFuncByName(element any) (any, error) {
	elemMap := element.(map[string]any)
	return elemMap["name"], nil
}

func testCreateSet(items []map[string]any) *schema.Set {
	return schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "test name",
				Required:    true,
			},
			"value": {
				Type:        schema.TypeString,
				Description: "test value",
				Required:    true,
			},
		},
	}), toArrayInterface(items))
}

func toArrayInterface(arrayOfMaps []map[string]any) []any {
	var result []any
	for _, c := range arrayOfMaps {
		result = append(result, c)
	}
	return result
}
