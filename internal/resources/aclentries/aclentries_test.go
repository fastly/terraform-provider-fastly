package aclentries

import (
	"context"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly/computeacls"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestExpandEntries(t *testing.T) {
	ctx := context.Background()

	t.Run("null map returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		result := expandEntries(ctx, types.MapNull(types.StringType), &diags)
		assert.Nil(t, result)
		assert.False(t, diags.HasError())
	})

	t.Run("populated map", func(t *testing.T) {
		m, d := types.MapValue(types.StringType, map[string]attr.Value{
			"192.0.2.0/24": types.StringValue("ALLOW"),
		})
		assert.False(t, d.HasError())

		var diags diag.Diagnostics
		result := expandEntries(ctx, m, &diags)
		assert.False(t, diags.HasError())
		assert.Equal(t, map[string]string{"192.0.2.0/24": "ALLOW"}, result)
	})
}

func TestFlattenEntries(t *testing.T) {
	var diags diag.Diagnostics
	remote := []computeacls.ComputeACLEntry{
		{Prefix: "192.0.2.0/24", Action: "ALLOW"},
		{Prefix: "198.51.100.0/24", Action: "BLOCK"},
	}

	result := flattenEntries(remote, &diags)
	assert.False(t, diags.HasError())

	var got map[string]string
	assert.False(t, result.ElementsAs(context.Background(), &got, false).HasError())
	assert.Equal(t, map[string]string{
		"192.0.2.0/24":    "ALLOW",
		"198.51.100.0/24": "BLOCK",
	}, got)
}

func TestBuildBatchEntries(t *testing.T) {
	tests := []struct {
		name     string
		old      map[string]string
		new      map[string]string
		manage   bool
		expected []*computeacls.BatchComputeACLEntry
	}{
		{
			name: "create only",
			old:  nil,
			new: map[string]string{
				"192.0.2.0/24": "ALLOW",
			},
			manage: true,
			expected: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("ALLOW"), Operation: new(createOperation)},
			},
		},
		{
			name: "update when prefix already exists",
			old: map[string]string{
				"192.0.2.0/24": "ALLOW",
			},
			new: map[string]string{
				"192.0.2.0/24": "BLOCK",
			},
			manage: true,
			expected: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("BLOCK"), Operation: new(updateOperation)},
			},
		},
		{
			name: "delete when managed and prefix removed",
			old: map[string]string{
				"192.0.2.0/24": "ALLOW",
			},
			new:    map[string]string{},
			manage: true,
			expected: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Operation: new(deleteOperation)},
			},
		},
		{
			name: "no delete when unmanaged",
			old: map[string]string{
				"192.0.2.0/24": "ALLOW",
			},
			new:      map[string]string{},
			manage:   false,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildBatchEntries(tt.old, tt.new, tt.manage)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestEntriesConverged(t *testing.T) {
	tests := []struct {
		name      string
		remote    []computeacls.ComputeACLEntry
		batch     []*computeacls.BatchComputeACLEntry
		converged bool
	}{
		{
			name:   "create not yet visible",
			remote: nil,
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("ALLOW"), Operation: new(createOperation)},
			},
			converged: false,
		},
		{
			name: "create visible",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "ALLOW"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("ALLOW"), Operation: new(createOperation)},
			},
			converged: true,
		},
		{
			name: "update not yet visible",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "ALLOW"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("BLOCK"), Operation: new(updateOperation)},
			},
			converged: false,
		},
		{
			name: "update visible",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "BLOCK"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("BLOCK"), Operation: new(updateOperation)},
			},
			converged: true,
		},
		{
			name: "delete not yet visible",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "ALLOW"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Operation: new(deleteOperation)},
			},
			converged: false,
		},
		{
			name:   "delete visible",
			remote: nil,
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Operation: new(deleteOperation)},
			},
			converged: true,
		},
		{
			name: "mixed operations all converged",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "198.51.100.0/24", Action: "ALLOW"},
				{Prefix: "203.0.113.0/24", Action: "BLOCK"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Operation: new(deleteOperation)},
				{Prefix: new("198.51.100.0/24"), Action: new("ALLOW"), Operation: new(updateOperation)},
				{Prefix: new("203.0.113.0/24"), Action: new("BLOCK"), Operation: new(createOperation)},
			},
			converged: true,
		},
		{
			name: "mixed operations one still pending",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "ALLOW"},
				{Prefix: "198.51.100.0/24", Action: "ALLOW"},
				{Prefix: "203.0.113.0/24", Action: "BLOCK"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Operation: new(deleteOperation)},
				{Prefix: new("198.51.100.0/24"), Action: new("ALLOW"), Operation: new(updateOperation)},
				{Prefix: new("203.0.113.0/24"), Action: new("BLOCK"), Operation: new(createOperation)},
			},
			converged: false,
		},
		{
			name: "unrelated remote entries are ignored",
			remote: []computeacls.ComputeACLEntry{
				{Prefix: "192.0.2.0/24", Action: "ALLOW"},
				{Prefix: "203.0.113.0/24", Action: "BLOCK"},
			},
			batch: []*computeacls.BatchComputeACLEntry{
				{Prefix: new("192.0.2.0/24"), Action: new("ALLOW"), Operation: new(createOperation)},
			},
			converged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.converged, entriesConverged(tt.remote, tt.batch))
		})
	}
}
