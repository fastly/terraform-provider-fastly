package aclentriescdn

import (
	"context"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestFlattenEntries(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		remote []*fastly.ACLEntry
		want   int
	}{
		{
			name: "multiple entries",
			remote: []*fastly.ACLEntry{
				{
					EntryID: stringPtr("entry1"),
					IP:      stringPtr("127.0.0.1"),
					Subnet:  intPtr(24),
					Negated: boolPtr(false),
					Comment: stringPtr("ACL Entry 1"),
				},
				{
					EntryID: stringPtr("entry2"),
					IP:      stringPtr("192.168.0.1"),
					Subnet:  intPtr(16),
					Negated: boolPtr(true),
					Comment: stringPtr("ACL Entry 2"),
				},
			},
			want: 2,
		},
		{
			name:   "empty entries",
			remote: []*fastly.ACLEntry{},
			want:   0,
		},
		{
			name:   "nil entries",
			remote: nil,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := flattenEntries(ctx, tt.remote, &diags)

			assert.False(t, diags.HasError())

			if tt.want == 0 {
				assert.True(t, result.IsNull() || len(result.Elements()) == 0)
			} else {
				assert.Equal(t, tt.want, len(result.Elements()))
			}

			if tt.name == "multiple entries" && len(result.Elements()) > 0 {
				elements := result.Elements()
				firstEntry := elements[0].(types.Object)
				attrs := firstEntry.Attributes()

				assert.Equal(t, "entry1", attrs["id"].(types.String).ValueString())
				assert.Equal(t, "127.0.0.1", attrs["ip"].(types.String).ValueString())
				assert.Equal(t, int64(24), attrs["subnet"].(types.Int64).ValueInt64())
				assert.Equal(t, false, attrs["negated"].(types.Bool).ValueBool())
				assert.Equal(t, "ACL Entry 1", attrs["comment"].(types.String).ValueString())
			}
		})
	}
}

func TestExpandEntries(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		input    types.Set
		wantLen  int
		wantErrs bool
	}{
		{
			name:     "null set",
			input:    types.SetNull(types.ObjectType{AttrTypes: entryAttrTypes()}),
			wantLen:  0,
			wantErrs: false,
		},
		{
			name:     "unknown set",
			input:    types.SetUnknown(types.ObjectType{AttrTypes: entryAttrTypes()}),
			wantLen:  0,
			wantErrs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := expandEntries(ctx, tt.input, &diags)

			assert.Equal(t, tt.wantErrs, diags.HasError())
			assert.Equal(t, tt.wantLen, len(result))
		})
	}
}

func TestBuildBatchACLEntry(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		entry EntryModel
		op    fastly.BatchOperation
		want  *fastly.BatchACLEntry
	}{
		{
			name: "complete entry",
			entry: EntryModel{
				ID:      types.StringValue("entry123"),
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(true),
				Comment: types.StringValue("test comment"),
			},
			op: fastly.CreateBatchOperation,
			want: &fastly.BatchACLEntry{
				Operation: func() *fastly.BatchOperation { op := fastly.CreateBatchOperation; return &op }(),
				EntryID:   func() *string { s := "entry123"; return &s }(),
				IP:        func() *string { s := "10.0.0.1"; return &s }(),
				Subnet:    func() *int { i := 24; return &i }(),
				Negated:   func() *fastly.Compatibool { b := fastly.Compatibool(true); return &b }(),
				Comment:   func() *string { s := "test comment"; return &s }(),
			},
		},
		{
			name: "minimal entry",
			entry: EntryModel{
				IP:      types.StringValue("192.168.1.1"),
				Negated: types.BoolValue(false),
			},
			op: fastly.CreateBatchOperation,
			want: &fastly.BatchACLEntry{
				Operation: func() *fastly.BatchOperation { op := fastly.CreateBatchOperation; return &op }(),
				IP:        func() *string { s := "192.168.1.1"; return &s }(),
				Negated:   func() *fastly.Compatibool { b := fastly.Compatibool(false); return &b }(),
			},
		},
		{
			name: "zero subnet",
			entry: EntryModel{
				IP:      types.StringValue("1.2.3.4"),
				Subnet:  types.Int64Value(0),
				Negated: types.BoolValue(false),
			},
			op: fastly.UpdateBatchOperation,
			want: &fastly.BatchACLEntry{
				Operation: func() *fastly.BatchOperation { op := fastly.UpdateBatchOperation; return &op }(),
				IP:        func() *string { s := "1.2.3.4"; return &s }(),
				Subnet:    func() *int { i := 0; return &i }(),
				Negated:   func() *fastly.Compatibool { b := fastly.Compatibool(false); return &b }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildBatchACLEntry(ctx, tt.entry, tt.op)

			assert.NotNil(t, got.Operation)
			assert.Equal(t, *tt.want.Operation, *got.Operation)

			if tt.want.EntryID != nil {
				assert.NotNil(t, got.EntryID)
				assert.Equal(t, *tt.want.EntryID, *got.EntryID)
			}

			if tt.want.IP != nil {
				assert.NotNil(t, got.IP)
				assert.Equal(t, *tt.want.IP, *got.IP)
			}

			if tt.want.Subnet != nil {
				assert.NotNil(t, got.Subnet)
				assert.Equal(t, *tt.want.Subnet, *got.Subnet)
			}

			if tt.want.Negated != nil {
				assert.NotNil(t, got.Negated)
				assert.Equal(t, *tt.want.Negated, *got.Negated)
			}

			if tt.want.Comment != nil {
				assert.NotNil(t, got.Comment)
				assert.Equal(t, *tt.want.Comment, *got.Comment)
			}
		})
	}
}

func TestEntriesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    EntryModel
		b    EntryModel
		want bool
	}{
		{
			name: "identical entries",
			a: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			b: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			want: true,
		},
		{
			name: "different IP",
			a: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			b: EntryModel{
				IP:      types.StringValue("10.0.0.2"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			want: false,
		},
		{
			name: "different subnet",
			a: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			b: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(16),
				Negated: types.BoolValue(false),
				Comment: types.StringValue("test"),
			},
			want: false,
		},
		{
			name: "null vs empty comment",
			a: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringNull(),
			},
			b: EntryModel{
				IP:      types.StringValue("10.0.0.1"),
				Subnet:  types.Int64Value(24),
				Negated: types.BoolValue(false),
				Comment: types.StringValue(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entriesEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEntryAttrTypes(t *testing.T) {
	attrs := entryAttrTypes()

	assert.Equal(t, 5, len(attrs))
	assert.Equal(t, types.StringType, attrs["id"])
	assert.Equal(t, types.StringType, attrs["ip"])
	assert.Equal(t, types.Int64Type, attrs["subnet"])
	assert.Equal(t, types.BoolType, attrs["negated"])
	assert.Equal(t, types.StringType, attrs["comment"])
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
