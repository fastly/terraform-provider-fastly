package vcl

import (
	"bytes"
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:    types.StringValue("main"),
		Content: types.StringValue("sub vcl_recv {\n#FASTLY recv\n  return(lookup);\n}\n"),
		Main:    types.BoolValue(true),
	}
}

func includedNestedModel() NestedModel {
	return NestedModel{
		Name:    types.StringValue("routing"),
		Content: types.StringValue("sub route_request {\n}\n"),
		Main:    types.BoolValue(false),
	}
}

func TestFlattenToNestedModel(t *testing.T) {
	t.Run("nil VCL returns zero model", func(t *testing.T) {
		model := FlattenToNestedModel(nil)

		assert.Equal(t, types.String{}, model.Name)
		assert.Equal(t, types.String{}, model.Content)
		assert.Equal(t, types.Bool{}, model.Main)
	})

	t.Run("full VCL", func(t *testing.T) {
		api := &fastly.VCL{
			Name:    fastly.ToPointer("main"),
			Content: fastly.ToPointer("sub vcl_recv {\n#FASTLY recv\n}\n"),
			Main:    fastly.ToPointer(true),
		}

		model := FlattenToNestedModel(api)

		assert.Equal(t, types.StringValue("main"), model.Name)
		assert.Equal(t, types.StringValue("sub vcl_recv {\n#FASTLY recv\n}\n"), model.Content)
		assert.Equal(t, types.BoolValue(true), model.Main)
	})

	t.Run("nil main defaults to false", func(t *testing.T) {
		api := &fastly.VCL{
			Name:    fastly.ToPointer("include"),
			Content: fastly.ToPointer("sub helper {}\n"),
		}

		model := FlattenToNestedModel(api)

		assert.Equal(t, types.BoolValue(false), model.Main)
	})
}

func TestFlatten(t *testing.T) {
	t.Run("nil VCL logs warning", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := tflogtest.RootLogger(context.Background(), &buf)
		m := &Model{}

		flatten(ctx, nil, m)

		assert.Equal(t, types.String{}, m.ID)
		assert.Equal(t, types.String{}, m.Service)
		assert.Equal(t, types.Int64{}, m.Version)
		assert.Equal(t, types.String{}, m.Name)
		assert.Equal(t, types.String{}, m.Content)
		assert.Equal(t, types.Bool{}, m.Main)

		entries, err := tflogtest.MultilineJSONDecode(&buf)
		require.NoError(t, err)
		require.NotEmpty(t, entries)

		foundWarning := false
		for _, entry := range entries {
			if entry["@level"] == "warn" && entry["@message"] == "flatten called with nil VCL" {
				foundWarning = true
				break
			}
		}
		assert.True(t, foundWarning)
	})

	t.Run("full VCL", func(t *testing.T) {
		ctx := context.Background()
		api := &fastly.VCL{
			ServiceID:      fastly.ToPointer("svc-123"),
			ServiceVersion: fastly.ToPointer(5),
			Name:           fastly.ToPointer("main"),
			Content:        fastly.ToPointer("sub vcl_recv {\n#FASTLY recv\n}\n"),
			Main:           fastly.ToPointer(true),
		}
		m := &Model{}

		flatten(ctx, api, m)

		assert.Equal(t, types.StringValue("svc-123-5-main"), m.ID)
		assert.Equal(t, types.StringValue("svc-123"), m.Service)
		assert.Equal(t, types.Int64Value(5), m.Version)
		assert.Equal(t, types.StringValue("main"), m.Name)
		assert.Equal(t, types.StringValue("sub vcl_recv {\n#FASTLY recv\n}\n"), m.Content)
		assert.Equal(t, types.BoolValue(true), m.Main)
	})
}

func TestBuildCreateInput(t *testing.T) {
	model := Model{
		Service: types.StringValue("svc-123"),
		Version: types.Int64Value(7),
		NestedModel: NestedModel{
			Name:    types.StringValue("main"),
			Content: types.StringValue("vcl content"),
			Main:    types.BoolValue(true),
		},
	}

	input := BuildCreateInput(model.Service.ValueString(), int(model.Version.ValueInt64()), model.NestedModel)

	assert.Equal(t, "svc-123", input.ServiceID)
	assert.Equal(t, 7, input.ServiceVersion)
	assert.Equal(t, "main", *input.Name)
	assert.Equal(t, "vcl content", *input.Content)
	assert.Equal(t, true, *input.Main)
}

func TestBuildUpdateInput(t *testing.T) {
	model := Model{
		Service: types.StringValue("svc-123"),
		Version: types.Int64Value(7),
		NestedModel: NestedModel{
			Name:    types.StringValue("main"),
			Content: types.StringValue("updated vcl content"),
			Main:    types.BoolValue(true),
		},
	}

	input := BuildUpdateInput(model.Service.ValueString(), int(model.Version.ValueInt64()), model.NestedModel)

	assert.Equal(t, "svc-123", input.ServiceID)
	assert.Equal(t, 7, input.ServiceVersion)
	assert.Equal(t, "main", input.Name)
	assert.Equal(t, "updated vcl content", *input.Content)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		vcls    []NestedModel
		wantErr string
	}{
		{
			name: "empty VCL list is valid",
			vcls: nil,
		},
		{
			name: "single main VCL is valid",
			vcls: []NestedModel{minimalNestedModel()},
		},
		{
			name: "main plus include is valid",
			vcls: []NestedModel{
				minimalNestedModel(),
				includedNestedModel(),
			},
		},
		{
			name: "duplicate names are invalid",
			vcls: []NestedModel{
				minimalNestedModel(),
				{
					Name:    types.StringValue("main"),
					Content: types.StringValue("different"),
					Main:    types.BoolValue(false),
				},
			},
			wantErr: `duplicate custom VCL name "main"`,
		},
		{
			name: "missing main is invalid",
			vcls: []NestedModel{
				includedNestedModel(),
			},
			wantErr: "one custom VCL file must have main = true",
		},
		{
			name: "multiple mains are invalid",
			vcls: []NestedModel{
				minimalNestedModel(),
				{
					Name:    types.StringValue("other-main"),
					Content: types.StringValue("sub vcl_recv {}\n"),
					Main:    types.BoolValue(true),
				},
			},
			wantErr: "only one custom VCL file can have main = true",
		},
		{
			name: "blank name is invalid",
			vcls: []NestedModel{
				{
					Name:    types.StringValue("  "),
					Content: types.StringValue("sub vcl_recv {}\n"),
					Main:    types.BoolValue(true),
				},
			},
			wantErr: "custom VCL name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.vcls)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []NestedModel
		b        []NestedModel
		expected bool
	}{
		{
			name:     "both empty",
			a:        []NestedModel{},
			b:        []NestedModel{},
			expected: true,
		},
		{
			name:     "identical single element",
			a:        []NestedModel{minimalNestedModel()},
			b:        []NestedModel{minimalNestedModel()},
			expected: true,
		},
		{
			name: "different order but same content",
			a: []NestedModel{
				{Name: types.StringValue("routing"), Content: types.StringValue("include"), Main: types.BoolValue(false)},
				{Name: types.StringValue("main"), Content: types.StringValue("main"), Main: types.BoolValue(true)},
			},
			b: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("main"), Main: types.BoolValue(true)},
				{Name: types.StringValue("routing"), Content: types.StringValue("include"), Main: types.BoolValue(false)},
			},
			expected: true,
		},
		{
			name: "different content",
			a: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("old"), Main: types.BoolValue(true)},
			},
			b: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("new"), Main: types.BoolValue(true)},
			},
			expected: false,
		},
		{
			name: "different main flag",
			a: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("content"), Main: types.BoolValue(true)},
			},
			b: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("content"), Main: types.BoolValue(false)},
			},
			expected: false,
		},
		{
			name: "different names",
			a: []NestedModel{
				{Name: types.StringValue("main"), Content: types.StringValue("content"), Main: types.BoolValue(true)},
			},
			b: []NestedModel{
				{Name: types.StringValue("other"), Content: types.StringValue("content"), Main: types.BoolValue(true)},
			},
			expected: false,
		},
		{
			name: "different lengths",
			a: []NestedModel{
				minimalNestedModel(),
			},
			b: []NestedModel{
				minimalNestedModel(),
				includedNestedModel(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Equal(tt.a, tt.b))
		})
	}
}

func TestContentEqual(t *testing.T) {
	assert.True(t, ContentEqual("sub vcl_recv {}\n", "sub vcl_recv {}\n\n"))
	assert.True(t, ContentEqual("sub vcl_recv {}", "sub vcl_recv {}\n"))
	assert.False(t, ContentEqual("sub vcl_recv { return(lookup); }\n", "sub vcl_recv { return(pass); }\n"))
}

func TestMatchOrderPreservePlanContent(t *testing.T) {
	remote := []NestedModel{
		{Name: types.StringValue("main"), Content: types.StringValue("api content\n\n"), Main: types.BoolValue(true)},
	}
	plan := []NestedModel{
		{Name: types.StringValue("main"), Content: types.StringValue("planned content\n"), Main: types.BoolValue(true)},
	}

	got := MatchOrderPreservePlanContent(remote, plan)

	require.Len(t, got, 1)
	assert.Equal(t, types.StringValue("planned content\n"), got[0].Content)
}

func TestMatchOrder(t *testing.T) {
	items := []NestedModel{
		{Name: types.StringValue("a"), Content: types.StringValue("a"), Main: types.BoolValue(false)},
		{Name: types.StringValue("b"), Content: types.StringValue("b"), Main: types.BoolValue(true)},
		{Name: types.StringValue("c"), Content: types.StringValue("c"), Main: types.BoolValue(false)},
	}
	order := []NestedModel{
		{Name: types.StringValue("c")},
		{Name: types.StringValue("a")},
	}

	got := MatchOrder(items, order)

	require.Len(t, got, 3)
	assert.Equal(t, "c", got[0].Name.ValueString())
	assert.Equal(t, "a", got[1].Name.ValueString())
	assert.Equal(t, "b", got[2].Name.ValueString())
}

func TestID(t *testing.T) {
	assert.Equal(t, "svc-123-5-main", ID("svc-123", 5, "main"))
}
