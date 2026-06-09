package domain

import (
	"bytes"
	"context"
	"testing"

	fastly "github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
)

func TestFlatten(t *testing.T) {
	t.Run("nil domain logs warning", func(t *testing.T) {
		var buf bytes.Buffer
		ctx := tflogtest.RootLogger(context.Background(), &buf)
		m := &Model{}
		flatten(ctx, nil, m)

		assert.Equal(t, types.String{}, m.ID)
		assert.Equal(t, types.String{}, m.Service)
		assert.Equal(t, types.Int64{}, m.Version)
		assert.Equal(t, types.String{}, m.Name)
		assert.Equal(t, types.String{}, m.Comment)

		entries, err := tflogtest.MultilineJSONDecode(&buf)
		assert.NoError(t, err)
		assert.NotEmpty(t, entries)
		foundWarning := false
		for _, entry := range entries {
			if entry["@level"] == "warn" && entry["@message"] == "flatten called with nil domain" {
				foundWarning = true
				break
			}
		}
		assert.True(t, foundWarning)
	})

	t.Run("full domain with comment", func(t *testing.T) {
		ctx := context.Background()
		domain := &fastly.Domain{
			ServiceID:      fastly.ToPointer("svc-123"),
			ServiceVersion: fastly.ToPointer(5),
			Name:           fastly.ToPointer("example.com"),
			Comment:        fastly.ToPointer("Test comment"),
		}
		m := &Model{}
		flatten(ctx, domain, m)

		assert.Equal(t, types.StringValue("svc-123-5-example.com"), m.ID)
		assert.Equal(t, types.StringValue("svc-123"), m.Service)
		assert.Equal(t, types.Int64Value(5), m.Version)
		assert.Equal(t, types.StringValue("example.com"), m.Name)
		assert.Equal(t, types.StringValue("Test comment"), m.Comment)
	})

	t.Run("minimal domain without comment", func(t *testing.T) {
		ctx := context.Background()
		domain := &fastly.Domain{
			ServiceID:      fastly.ToPointer("svc-456"),
			ServiceVersion: fastly.ToPointer(1),
			Name:           fastly.ToPointer("minimal.com"),
		}
		m := &Model{}
		flatten(ctx, domain, m)

		assert.Equal(t, types.StringValue("svc-456-1-minimal.com"), m.ID)
		assert.Equal(t, types.StringValue("svc-456"), m.Service)
		assert.Equal(t, types.Int64Value(1), m.Version)
		assert.Equal(t, types.StringValue("minimal.com"), m.Name)
		assert.True(t, m.Comment.IsNull())
	})

	t.Run("empty and whitespace comments", func(t *testing.T) {
		ctx := context.Background()

		emptyDomain := &fastly.Domain{
			ServiceID:      fastly.ToPointer("svc-1"),
			ServiceVersion: fastly.ToPointer(1),
			Name:           fastly.ToPointer("empty.com"),
			Comment:        fastly.ToPointer(""),
		}
		m1 := &Model{}
		flatten(ctx, emptyDomain, m1)
		assert.True(t, m1.Comment.IsNull())

		whitespaceDomain := &fastly.Domain{
			ServiceID:      fastly.ToPointer("svc-2"),
			ServiceVersion: fastly.ToPointer(2),
			Name:           fastly.ToPointer("whitespace.com"),
			Comment:        fastly.ToPointer("   "),
		}
		m2 := &Model{}
		flatten(ctx, whitespaceDomain, m2)
		assert.Equal(t, types.StringValue("   "), m2.Comment)
	})
}

func TestExpandCreate(t *testing.T) {
	t.Run("minimal domain", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-123"),
			Version: types.Int64Value(5),
			Name:    types.StringValue("example.com"),
			Comment: types.StringNull(),
		}
		input := expandCreate(model)

		assert.Equal(t, "svc-123", input.ServiceID)
		assert.Equal(t, 5, input.ServiceVersion)
		assert.Equal(t, "example.com", *input.Name)
		assert.Nil(t, input.Comment)
	})

	t.Run("domain with comment", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-456"),
			Version: types.Int64Value(1),
			Name:    types.StringValue("test.com"),
			Comment: types.StringValue("Test comment"),
		}
		input := expandCreate(model)

		assert.Equal(t, "svc-456", input.ServiceID)
		assert.Equal(t, 1, input.ServiceVersion)
		assert.Equal(t, "test.com", *input.Name)
		assert.Equal(t, "Test comment", *input.Comment)
	})

	t.Run("comment handling", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-789"),
			Version: types.Int64Value(3),
			Name:    types.StringValue("test.com"),
		}

		model.Comment = types.StringValue("")
		assert.Nil(t, expandCreate(model).Comment)

		model.Comment = types.StringValue("   ")
		assert.Equal(t, "   ", *expandCreate(model).Comment)
	})
}

func TestExpandUpdate(t *testing.T) {
	t.Run("minimal domain", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-123"),
			Version: types.Int64Value(5),
			Name:    types.StringValue("example.com"),
			Comment: types.StringNull(),
		}
		input := expandUpdate(model)

		assert.Equal(t, "svc-123", input.ServiceID)
		assert.Equal(t, 5, input.ServiceVersion)
		assert.Equal(t, "example.com", input.Name)
		assert.Nil(t, input.Comment)
	})

	t.Run("domain with comment", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-456"),
			Version: types.Int64Value(1),
			Name:    types.StringValue("test.com"),
			Comment: types.StringValue("Updated comment"),
		}
		input := expandUpdate(model)

		assert.Equal(t, "svc-456", input.ServiceID)
		assert.Equal(t, 1, input.ServiceVersion)
		assert.Equal(t, "test.com", input.Name)
		assert.Equal(t, "Updated comment", *input.Comment)
	})

	t.Run("comment handling", func(t *testing.T) {
		model := Model{
			Service: types.StringValue("svc-789"),
			Version: types.Int64Value(3),
			Name:    types.StringValue("test.com"),
		}

		model.Comment = types.StringNull()
		assert.Nil(t, expandUpdate(model).Comment)

		model.Comment = types.StringValue("")
		assert.Nil(t, expandUpdate(model).Comment)
	})
}

func TestIDGeneration(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		svc, name, expectedID string
		ver                   int
	}{
		{"svc1", "example.com", "svc1-1-example.com", 1},
		{"svc2", "api.example.com", "svc2-10-api.example.com", 10},
		{"svc3", "*.example.com", "svc3-5-*.example.com", 5},
	}

	for _, c := range cases {
		domain := &fastly.Domain{
			ServiceID:      fastly.ToPointer(c.svc),
			ServiceVersion: fastly.ToPointer(c.ver),
			Name:           fastly.ToPointer(c.name),
		}
		m := &Model{}
		flatten(ctx, domain, m)
		assert.Equal(t, types.StringValue(c.expectedID), m.ID)
	}
}
