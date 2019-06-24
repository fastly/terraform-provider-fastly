package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"reflect"
	"testing"
)

func TestResourceFastlyFlattenDictionaryItems(t *testing.T) {
	cases := []struct {
		remote []*gofastly.ACLEntry
		local  map[string]string
	}{
		{
			remote: []*gofastly.ACLEntry{
				{
					ServiceID:    "service-id",
					ACLID: "1234567890",
				},
				{
					ServiceID:    "service-id",
					ACLID: "1234567890",
				},
			},
			local: map[string]string{

			},
		},
	}

	for _, c := range cases {
		out := flattenAclEntries(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}