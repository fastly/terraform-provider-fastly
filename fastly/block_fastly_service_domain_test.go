package fastly

import (
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
)

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]any
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    gofastly.ToPointer("test.notexample.com"),
					Comment: gofastly.ToPointer("not comment"),
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
					Name: gofastly.ToPointer("test.notexample.com"),
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
