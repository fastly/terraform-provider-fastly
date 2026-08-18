package director

import (
	"context"
	"testing"

	"github.com/fastly/terraform-provider-fastly/internal/resources/backend"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func minimalNestedModel() NestedModel {
	return NestedModel{
		Name:     types.StringValue("director"),
		Backends: setValue("origin1"),
		Comment:  types.StringValue(""),
		Quorum:   types.Int64Value(75),
		Retries:  types.Int64Value(5),
		Shield:   types.StringValue(""),
		Type:     types.StringValue("random"),
	}
}

func fullNestedModel() NestedModel {
	return NestedModel{
		Name:     types.StringValue("director"),
		Backends: setValue("origin1", "origin2"),
		Comment:  types.StringValue("a comment"),
		Quorum:   types.Int64Value(30),
		Retries:  types.Int64Value(10),
		Shield:   types.StringValue("sjc-ca-us"),
		Type:     types.StringValue("hash"),
	}
}

func setValue(values ...string) types.Set {
	return stringSliceToSet(values)
}

func TestToModel(t *testing.T) {
	directorType := fastly.DirectorTypeHash
	api := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin2", "origin1"},
		Comment:  new("a comment"),
		Quorum:   new(30),
		Retries:  new(10),
		Shield:   new("sjc-ca-us"),
		Type:     &directorType,
	}

	result := ops{}.ToModel(api)

	assert.True(t, fullNestedModel().ModelsEqual(result))
}

func TestOpsEqual(t *testing.T) {
	directorType := fastly.DirectorTypeHash
	remote := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin1", "origin2"},
		Comment:  new("a comment"),
		Quorum:   new(30),
		Retries:  new(10),
		Shield:   new("sjc-ca-us"),
		Type:     &directorType,
	}

	assert.True(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestOpsEqual_backendOrderIndependent(t *testing.T) {
	directorType := fastly.DirectorTypeRandom
	remote := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin1"},
		Comment:  new(""),
		Quorum:   new(75),
		Retries:  new(5),
		Shield:   new(""),
		Type:     &directorType,
	}

	desired := minimalNestedModel()
	desired.Backends = setValue("origin1")

	assert.True(t, ops{}.Equal(desired, remote))
}

func TestOpsEqual_mismatch(t *testing.T) {
	directorType := fastly.DirectorTypeRandom
	remote := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin1"},
		Comment:  new(""),
		Quorum:   new(75),
		Retries:  new(5),
		Shield:   new(""),
		Type:     &directorType,
	}

	assert.False(t, ops{}.Equal(fullNestedModel(), remote))
}

func TestDirectorTypePointer(t *testing.T) {
	tests := []struct {
		name     string
		value    types.String
		expected fastly.DirectorType
	}{
		{name: "random", value: types.StringValue("random"), expected: fastly.DirectorTypeRandom},
		{name: "hash", value: types.StringValue("hash"), expected: fastly.DirectorTypeHash},
		{name: "client", value: types.StringValue("client"), expected: fastly.DirectorTypeClient},
		{name: "round_robin", value: types.StringValue("round_robin"), expected: fastly.DirectorTypeRoundRobin},
		{name: "numeric alias 1", value: types.StringValue("1"), expected: fastly.DirectorTypeRandom},
		{name: "numeric alias 3", value: types.StringValue("3"), expected: fastly.DirectorTypeHash},
		{name: "numeric alias 4", value: types.StringValue("4"), expected: fastly.DirectorTypeClient},
		{name: "numeric alias 2", value: types.StringValue("2"), expected: fastly.DirectorTypeRoundRobin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := directorTypePointer(tt.value)
			if assert.NotNil(t, result) {
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func TestDirectorTypePointer_unrecognizedPanics(t *testing.T) {
	assert.Panics(t, func() {
		directorTypePointer(types.StringValue("bogus"))
	})
}

func TestToModel_roundRobin(t *testing.T) {
	directorType := fastly.DirectorTypeRoundRobin
	api := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin1"},
		Comment:  new(""),
		Quorum:   new(75),
		Retries:  new(5),
		Shield:   new(""),
		Type:     &directorType,
	}

	result := ops{}.ToModel(api)

	assert.Equal(t, "round_robin", result.Type.ValueString())
}

func TestDirectorTypeCanonical(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
	}{
		{alias: "random", expected: "random"},
		{alias: "1", expected: "random"},
		{alias: "hash", expected: "hash"},
		{alias: "3", expected: "hash"},
		{alias: "client", expected: "client"},
		{alias: "4", expected: "client"},
		{alias: "round_robin", expected: "round_robin"},
		{alias: "2", expected: "round_robin"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			result, ok := directorTypeCanonical(tt.alias)
			if assert.True(t, ok) {
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	t.Run("unrecognized", func(t *testing.T) {
		_, ok := directorTypeCanonical("bogus")
		assert.False(t, ok)
	})
}

func TestSetToStringSlice(t *testing.T) {
	result := setToStringSlice(setValue("b", "a"))
	assert.ElementsMatch(t, []string{"a", "b"}, result)
}

func TestSetToStringSlice_skipsUnknownAndNull(t *testing.T) {
	s := types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("origin1"),
		types.StringUnknown(),
		types.StringNull(),
	})

	result := setToStringSlice(s)
	assert.Equal(t, []string{"origin1"}, result)
}

func TestMetadataFieldsEqual(t *testing.T) {
	directorType := fastly.DirectorTypeHash
	remote := &fastly.Director{
		Name:     new("director"),
		Backends: []string{"origin1"},
		Comment:  new("a comment"),
		Quorum:   new(30),
		Retries:  new(10),
		Shield:   new("sjc-ca-us"),
		Type:     &directorType,
	}

	t.Run("equal ignoring backends", func(t *testing.T) {
		desired := fullNestedModel()
		desired.Backends = setValue("some-other-backend")
		assert.True(t, ops{}.metadataFieldsEqual(desired, remote))
	})

	t.Run("differs on a non-backend field", func(t *testing.T) {
		desired := fullNestedModel()
		desired.Comment = types.StringValue("different")
		assert.False(t, ops{}.metadataFieldsEqual(desired, remote))
	})
}

func directorObjectType() types.ObjectType {
	attrTypes := make(map[string]attr.Type, len(CommonAttributes()))
	for name, a := range CommonAttributes() {
		attrTypes[name] = a.GetType()
	}
	return types.ObjectType{AttrTypes: attrTypes}
}

func directorObjectValue(t *testing.T, name string, typ types.String) types.Object {
	t.Helper()
	m := minimalNestedModel()
	m.Name = types.StringValue(name)
	m.Type = typ

	objType := directorObjectType()
	obj, diags := types.ObjectValueFrom(context.Background(), objType.AttrTypes, m)
	if diags.HasError() {
		t.Fatalf("building object value: %v", diags)
	}
	return obj
}

func directorListValue(t *testing.T, objs ...types.Object) types.List {
	t.Helper()
	elems := make([]attr.Value, len(objs))
	for i, o := range objs {
		elems[i] = o
	}
	list, diags := types.ListValue(directorObjectType(), elems)
	if diags.HasError() {
		t.Fatalf("building list value: %v", diags)
	}
	return list
}

func directorTypeOf(t *testing.T, list types.List, index int) types.String {
	t.Helper()
	elems := list.Elements()
	if index >= len(elems) {
		t.Fatalf("index %d out of range (len %d)", index, len(elems))
	}
	obj, ok := elems[index].(types.Object)
	if !ok {
		t.Fatalf("element %d is not an object", index)
	}
	v, ok := obj.Attributes()["type"].(types.String)
	if !ok {
		t.Fatalf("element %d has no string type attribute", index)
	}
	return v
}

func TestTypeStickyDefault(t *testing.T) {
	t.Run("create - config omitted defaults", func(t *testing.T) {
		configObj := directorObjectValue(t, "a", types.StringNull())
		planObj := directorObjectValue(t, "a", types.StringNull())

		req := planmodifier.ListRequest{
			ConfigValue: directorListValue(t, configObj),
			StateValue:  types.ListNull(directorObjectType()),
			PlanValue:   directorListValue(t, planObj),
		}
		resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

		typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

		assert.Equal(t, types.StringValue(DefaultType), directorTypeOf(t, resp.PlanValue, 0))
	})

	t.Run("update - config omitted preserves round_robin by name", func(t *testing.T) {
		configObj := directorObjectValue(t, "a", types.StringNull())
		planObj := directorObjectValue(t, "a", types.StringNull())
		stateObj := directorObjectValue(t, "a", types.StringValue("round_robin"))

		req := planmodifier.ListRequest{
			ConfigValue: directorListValue(t, configObj),
			StateValue:  directorListValue(t, stateObj),
			PlanValue:   directorListValue(t, planObj),
		}
		resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

		typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

		assert.Equal(t, types.StringValue("round_robin"), directorTypeOf(t, resp.PlanValue, 0))
	})

	t.Run("update - config omitted resets a non-round_robin type to the default", func(t *testing.T) {
		// Regression test: type has a schema-level Default of 1 (random) in the legacy SDKv2
		// provider on main, so dropping an explicit `type = "hash"` from config resets the
		// director to random there. round_robin is the only exception (see the typeStickyDefault
		// doc comment) - every other prior value must reset on omit, not stick.
		for _, prior := range []string{"random", "hash", "client"} {
			t.Run(prior, func(t *testing.T) {
				configObj := directorObjectValue(t, "a", types.StringNull())
				planObj := directorObjectValue(t, "a", types.StringNull())
				stateObj := directorObjectValue(t, "a", types.StringValue(prior))

				req := planmodifier.ListRequest{
					ConfigValue: directorListValue(t, configObj),
					StateValue:  directorListValue(t, stateObj),
					PlanValue:   directorListValue(t, planObj),
				}
				resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

				typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

				assert.Equal(t, types.StringValue(DefaultType), directorTypeOf(t, resp.PlanValue, 0))
			})
		}
	})

	t.Run("config set - left to the configured value", func(t *testing.T) {
		configObj := directorObjectValue(t, "a", types.StringValue("hash"))
		planObj := directorObjectValue(t, "a", types.StringValue("hash"))
		stateObj := directorObjectValue(t, "a", types.StringValue("random"))

		req := planmodifier.ListRequest{
			ConfigValue: directorListValue(t, configObj),
			StateValue:  directorListValue(t, stateObj),
			PlanValue:   directorListValue(t, planObj),
		}
		resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

		typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

		assert.Equal(t, types.StringValue("hash"), directorTypeOf(t, resp.PlanValue, 0))
	})

	t.Run("config set to a numeric alias - normalized to the friendly name", func(t *testing.T) {
		// A configured "3" must be normalized to "hash" in the plan itself, not just on the next
		// refresh: Create/Read will return "hash" from the API, and if the plan still said "3"
		// the framework's post-apply consistency check would fail ("Provider produced
		// inconsistent result after apply").
		configObj := directorObjectValue(t, "a", types.StringValue("3"))
		planObj := directorObjectValue(t, "a", types.StringValue("3"))

		req := planmodifier.ListRequest{
			ConfigValue: directorListValue(t, configObj),
			StateValue:  types.ListNull(directorObjectType()),
			PlanValue:   directorListValue(t, planObj),
		}
		resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

		typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

		assert.Equal(t, types.StringValue("hash"), directorTypeOf(t, resp.PlanValue, 0))
	})

	t.Run("inserting a director ahead of an existing one matches by name, not position", func(t *testing.T) {
		// Regression test: state has [a(round_robin)] at index 0 - only round_robin sticks on
		// omit, so it's the only prior value that would expose a positional mismatch here. Config
		// inserts a new director "c" (no type) ahead of "a" (also no type), so at index 0 the
		// plan-modifier machinery would naively pair against state's index-0 element
		// ("a", round_robin) if this modifier matched positionally instead of by name - wrongly
		// making the new directorC round_robin, and resetting directorA to the default instead of
		// preserving its round_robin status.
		stateObj := directorObjectValue(t, "a", types.StringValue("round_robin"))

		configC := directorObjectValue(t, "c", types.StringNull())
		configA := directorObjectValue(t, "a", types.StringNull())
		planC := directorObjectValue(t, "c", types.StringNull())
		planA := directorObjectValue(t, "a", types.StringNull())

		req := planmodifier.ListRequest{
			ConfigValue: directorListValue(t, configC, configA),
			StateValue:  directorListValue(t, stateObj),
			PlanValue:   directorListValue(t, planC, planA),
		}
		resp := &planmodifier.ListResponse{PlanValue: req.PlanValue}

		typeStickyDefault{}.PlanModifyList(context.Background(), req, resp)

		assert.Equal(t, types.StringValue(DefaultType), directorTypeOf(t, resp.PlanValue, 0), "new director c should default, not inherit a's round_robin type")
		assert.Equal(t, types.StringValue("round_robin"), directorTypeOf(t, resp.PlanValue, 1), "existing director a should keep its own round_robin type by name")
	})
}

func TestStringSliceToSet(t *testing.T) {
	t.Run("sorted deterministically", func(t *testing.T) {
		result := stringSliceToSet([]string{"b", "a"})
		assert.Equal(t, []string{"a", "b"}, setToStringSlice(result))
	})

	t.Run("empty", func(t *testing.T) {
		result := stringSliceToSet(nil)
		assert.False(t, result.IsNull())
		assert.Empty(t, result.Elements())
	})

	t.Run("deduplicates instead of panicking", func(t *testing.T) {
		var result types.Set
		assert.NotPanics(t, func() {
			result = stringSliceToSet([]string{"a", "b", "a"})
		})
		assert.Equal(t, []string{"a", "b"}, setToStringSlice(result))
	})
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
			name:     "same content",
			a:        []NestedModel{fullNestedModel()},
			b:        []NestedModel{fullNestedModel()},
			expected: true,
		},
		{
			name: "different lengths",
			a:    []NestedModel{minimalNestedModel()},
			b: []NestedModel{
				minimalNestedModel(),
				fullNestedModel(),
			},
			expected: false,
		},
		{
			name:     "different content",
			a:        []NestedModel{minimalNestedModel()},
			b:        []NestedModel{fullNestedModel()},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Equal(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchOrder(t *testing.T) {
	a := NestedModel{Name: types.StringValue("a")}
	b := NestedModel{Name: types.StringValue("b")}

	result := MatchOrder([]NestedModel{b, a}, []NestedModel{a, b})

	assert.Equal(t, []NestedModel{a, b}, result)
}

func TestValidateConfig(t *testing.T) {
	t.Run("unique names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringValue("first")
		second := minimalNestedModel()
		second.Name = types.StringValue("second")

		assert.NoError(t, ValidateConfig([]NestedModel{first, second}))
	})

	t.Run("duplicate names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringValue("duplicate")
		second := minimalNestedModel()
		second.Name = types.StringValue("duplicate")

		err := ValidateConfig([]NestedModel{first, second})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), `"duplicate"`)
		}
	})

	t.Run("skips unknown names", func(t *testing.T) {
		first := minimalNestedModel()
		first.Name = types.StringUnknown()
		second := minimalNestedModel()
		second.Name = types.StringUnknown()

		assert.NoError(t, ValidateConfig([]NestedModel{first, second}))
	})
}

func TestValidateBackendReferences(t *testing.T) {
	be := func(name string) backend.NestedModel {
		return backend.NestedModel{Name: types.StringValue(name)}
	}

	t.Run("references configured backends", func(t *testing.T) {
		item := minimalNestedModel()
		item.Backends = setValue("origin1")

		assert.NoError(t, ValidateBackendReferences([]NestedModel{item}, []backend.NestedModel{be("origin1")}))
	})

	t.Run("references a backend that isn't configured", func(t *testing.T) {
		item := minimalNestedModel()
		item.Name = types.StringValue("director")
		item.Backends = setValue("missing-backend")

		err := ValidateBackendReferences([]NestedModel{item}, []backend.NestedModel{be("origin1")})
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), `"director"`)
			assert.Contains(t, err.Error(), `"missing-backend"`)
		}
	})

	t.Run("skips unknown backends", func(t *testing.T) {
		item := minimalNestedModel()
		item.Backends = types.SetUnknown(types.StringType)

		assert.NoError(t, ValidateBackendReferences([]NestedModel{item}, nil))
	})
}
