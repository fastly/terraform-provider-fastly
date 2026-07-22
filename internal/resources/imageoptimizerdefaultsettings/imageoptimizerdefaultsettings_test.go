package imageoptimizerdefaultsettings

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func fullNestedModel() NestedModel {
	return NestedModel{
		AllowVideo:   types.BoolValue(true),
		JpegQuality:  types.Int64Value(90),
		JpegType:     types.StringValue("progressive"),
		ResizeFilter: types.StringValue("bicubic"),
		Upscale:      types.BoolValue(true),
		Webp:         types.BoolValue(true),
		WebpQuality:  types.Int64Value(80),
	}
}

func TestFlattenToNestedModel(t *testing.T) {
	t.Run("nil settings", func(t *testing.T) {
		m := FlattenToNestedModel(nil)
		assert.Equal(t, NestedModel{}, m)
	})

	t.Run("full settings", func(t *testing.T) {
		s := &fastly.ImageOptimizerDefaultSettings{
			AllowVideo:   true,
			JpegQuality:  90,
			JpegType:     "progressive",
			ResizeFilter: "bicubic",
			Upscale:      true,
			Webp:         true,
			WebpQuality:  80,
		}
		m := FlattenToNestedModel(s)
		assert.Equal(t, fullNestedModel(), m)
	})
}

func TestDefaultNestedModel(t *testing.T) {
	d := defaultNestedModel()
	assert.Equal(t, types.BoolValue(false), d.AllowVideo)
	assert.Equal(t, types.Int64Value(85), d.JpegQuality)
	assert.Equal(t, types.StringValue("auto"), d.JpegType)
	assert.Equal(t, types.StringValue("lanczos3"), d.ResizeFilter)
	assert.Equal(t, types.BoolValue(false), d.Upscale)
	assert.Equal(t, types.BoolValue(false), d.Webp)
	assert.Equal(t, types.Int64Value(85), d.WebpQuality)
}

func TestModelsEqual(t *testing.T) {
	a := fullNestedModel()
	b := fullNestedModel()
	assert.True(t, a.ModelsEqual(b))

	b.Webp = types.BoolValue(false)
	assert.False(t, a.ModelsEqual(b))
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []NestedModel
		b        []NestedModel
		expected bool
	}{
		{name: "both empty", a: nil, b: nil, expected: true},
		{name: "different lengths", a: []NestedModel{fullNestedModel()}, b: nil, expected: false},
		{name: "equal single", a: []NestedModel{fullNestedModel()}, b: []NestedModel{fullNestedModel()}, expected: true},
		{
			name:     "different single",
			a:        []NestedModel{fullNestedModel()},
			b:        []NestedModel{defaultNestedModel()},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Equal(tt.a, tt.b))
		})
	}
}

func TestParseResizeFilter(t *testing.T) {
	for _, v := range ResizeFilters {
		t.Run(v, func(t *testing.T) {
			rf, err := parseResizeFilter(v)
			assert.NoError(t, err)
			assert.Equal(t, v, rf.String())
		})
	}

	t.Run("invalid", func(t *testing.T) {
		_, err := parseResizeFilter("invalid")
		assert.Error(t, err)
	})
}

func TestParseJpegType(t *testing.T) {
	for _, v := range JpegTypes {
		t.Run(v, func(t *testing.T) {
			jt, err := parseJpegType(v)
			assert.NoError(t, err)
			assert.Equal(t, v, jt.String())
		})
	}

	t.Run("invalid", func(t *testing.T) {
		_, err := parseJpegType("invalid")
		assert.Error(t, err)
	})
}

func TestBuildUpdateInput(t *testing.T) {
	t.Run("valid model", func(t *testing.T) {
		input, err := BuildUpdateInput("svc-123", 5, fullNestedModel())
		assert.NoError(t, err)
		assert.Equal(t, "svc-123", input.ServiceID)
		assert.Equal(t, 5, input.ServiceVersion)
		assert.Equal(t, true, *input.AllowVideo)
		assert.Equal(t, 90, *input.JpegQuality)
		assert.Equal(t, "progressive", input.JpegType.String())
		assert.Equal(t, "bicubic", input.ResizeFilter.String())
		assert.Equal(t, true, *input.Upscale)
		assert.Equal(t, true, *input.Webp)
		assert.Equal(t, 80, *input.WebpQuality)
	})

	t.Run("invalid jpeg_type", func(t *testing.T) {
		m := fullNestedModel()
		m.JpegType = types.StringValue("invalid")
		_, err := BuildUpdateInput("svc-123", 5, m)
		assert.Error(t, err)
	})

	t.Run("invalid resize_filter", func(t *testing.T) {
		m := fullNestedModel()
		m.ResizeFilter = types.StringValue("invalid")
		_, err := BuildUpdateInput("svc-123", 5, m)
		assert.Error(t, err)
	})
}
