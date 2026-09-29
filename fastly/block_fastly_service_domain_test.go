package fastly

import (
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v17/fastly"
)

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    new("test.notexample.com"),
					Comment: new("not comment"),
				},
			},
			local: []map[string]any{
				{
					"name":    "test.notexample.com",
					"comment": "not comment",
				},
			},
		},
		{
			remote: []*gofastly.Domain{
				{
					Name: new("test.notexample.com"),
				},
			},
			local: []map[string]any{
				{
					"name": "test.notexample.com",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDomains(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}
