package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetDiff_Diff(t *testing.T) {
	cases := []struct {
		name               string
		keyFunc            KeyFunc
		oldElements        []map[string]interface{}
		newElements        []map[string]interface{}
		expectedAdded      []map[string]interface{}
		expectedModified   []map[string]interface{}
		expectedDeleted    []map[string]interface{}
		expectedUnmodified []map[string]interface{}
		expectedError      bool
	}{
		{
			name: "should return the correct diff",
			oldElements: []map[string]interface{}{
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
			newElements: []map[string]interface{}{
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
			expectedAdded: []map[string]interface{}{
				{
					"name":  "name-c",
					"value": "value-c",
				},
			},
			expectedModified: []map[string]interface{}{
				{
					"name":  "name-a",
					"value": "value-a-new",
				},
			},
			expectedDeleted: []map[string]interface{}{
				{
					"name":  "b",
					"value": "value-b",
				},
			},
			expectedUnmodified: []map[string]interface{}{
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
			oldElements: []map[string]interface{}{
				{},
			},
			newElements: []map[string]interface{}{
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

func testKeyFuncByName(element interface{}) (interface{}, error) {
	elemMap := element.(map[string]interface{})
	return elemMap["name"], nil
}

func testCreateSet(items []map[string]interface{}) *schema.Set {
	return schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}), toArrayInterface(items))
}

func toArrayInterface(arrayOfMaps []map[string]interface{}) []interface{} {
	var result []interface{}
	for _, c := range arrayOfMaps {
		result = append(result, c)
	}
	return result
}
