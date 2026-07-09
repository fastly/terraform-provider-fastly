package acl

import (
	"testing"

	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		acl      *computeacls.ComputeACL
		initial  Model
		expected Model
	}{
		{
			name:     "nil ACL leaves model untouched",
			acl:      nil,
			initial:  Model{ID: types.StringValue("unchanged"), Name: types.StringValue("unchanged")},
			expected: Model{ID: types.StringValue("unchanged"), Name: types.StringValue("unchanged")},
		},
		{
			name: "populated ACL",
			acl: &computeacls.ComputeACL{
				ComputeACLID: "acl_abc123",
				Name:         "my_acl",
			},
			initial: Model{},
			expected: Model{
				ID:   types.StringValue("acl_abc123"),
				Name: types.StringValue("my_acl"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.initial
			flatten(&m, tt.acl)
			assert.Equal(t, tt.expected, m)
		})
	}
}
