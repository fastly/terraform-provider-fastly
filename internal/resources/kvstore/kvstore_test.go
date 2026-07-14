package kvstore

import (
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		store    *fastly.KVStore
		initial  Model
		expected Model
	}{
		{
			name:     "nil store leaves model untouched",
			store:    nil,
			initial:  Model{ID: types.StringValue("unchanged"), Name: types.StringValue("unchanged")},
			expected: Model{ID: types.StringValue("unchanged"), Name: types.StringValue("unchanged")},
		},
		{
			name: "populated store",
			store: &fastly.KVStore{
				StoreID: "store_abc123",
				Name:    "my_store",
			},
			initial: Model{Location: types.StringValue("US")},
			expected: Model{
				ID:       types.StringValue("store_abc123"),
				Name:     types.StringValue("my_store"),
				Location: types.StringValue("US"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.initial
			flatten(&m, tt.store)
			assert.Equal(t, tt.expected, m)
		})
	}
}
