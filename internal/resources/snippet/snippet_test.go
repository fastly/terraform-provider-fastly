package snippet

import (
	"testing"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenToNestedModel(t *testing.T) {
	api := &fastly.Snippet{
		Name:     fastly.ToPointer("recv_test"),
		Type:     fastly.ToPointer(fastly.SnippetTypeRecv),
		Priority: fastly.ToPointer("110"),
		Content:  fastly.ToPointer(`set req.http.X-Test = "true";`),
	}

	got, err := FlattenToNestedModel(api)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got.Name.ValueString() != "recv_test" {
		t.Fatalf("name mismatch: %q", got.Name.ValueString())
	}
	if got.Type.ValueString() != "recv" {
		t.Fatalf("type mismatch: %q", got.Type.ValueString())
	}
	if got.Priority.ValueInt64() != 110 {
		t.Fatalf("priority mismatch: %d", got.Priority.ValueInt64())
	}
	if got.Content.ValueString() != `set req.http.X-Test = "true";` {
		t.Fatalf("content mismatch: %q", got.Content.ValueString())
	}
}

func TestFlattenToNestedModelInvalidPriority(t *testing.T) {
	api := &fastly.Snippet{
		Name:     fastly.ToPointer("recv_test"),
		Type:     fastly.ToPointer(fastly.SnippetTypeRecv),
		Priority: fastly.ToPointer("invalid"),
		Content:  fastly.ToPointer(`set req.http.X-Test = "true";`),
	}

	_, err := FlattenToNestedModel(api)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildCreateInput(t *testing.T) {
	input := BuildCreateInput("service123", 2, NestedModel{
		Name:     types.StringValue("deliver_test"),
		Type:     types.StringValue("deliver"),
		Priority: types.Int64Value(50),
		Content:  types.StringValue(`set resp.http.X-Test = "true";`),
	})

	if input.ServiceID != "service123" {
		t.Fatalf("service id mismatch: %q", input.ServiceID)
	}
	if input.ServiceVersion != 2 {
		t.Fatalf("version mismatch: %d", input.ServiceVersion)
	}
	if fastly.ToValue(input.Name) != "deliver_test" {
		t.Fatalf("name mismatch: %q", fastly.ToValue(input.Name))
	}
	if fastly.ToValue(input.Dynamic) != 0 {
		t.Fatalf("dynamic mismatch: %d", fastly.ToValue(input.Dynamic))
	}
	if fastly.ToValue(input.Priority) != "50" {
		t.Fatalf("priority mismatch: %q", fastly.ToValue(input.Priority))
	}
	if fastly.ToValue(input.Type) != fastly.SnippetTypeDeliver {
		t.Fatalf("type mismatch: %q", fastly.ToValue(input.Type))
	}
}

func TestBuildUpdateInput(t *testing.T) {
	input := BuildUpdateInput("service123", 2, NestedModel{
		Name:     types.StringValue("deliver_test"),
		Type:     types.StringValue("deliver"),
		Priority: types.Int64Value(50),
		Content:  types.StringValue(`set resp.http.X-Test = "true";`),
	})

	if input.ServiceID != "service123" {
		t.Fatalf("service id mismatch: %q", input.ServiceID)
	}
	if input.ServiceVersion != 2 {
		t.Fatalf("version mismatch: %d", input.ServiceVersion)
	}
	if input.Name != "deliver_test" {
		t.Fatalf("name mismatch: %q", input.Name)
	}
	if fastly.ToValue(input.NewName) != "deliver_test" {
		t.Fatalf("new name mismatch: %q", fastly.ToValue(input.NewName))
	}
	if fastly.ToValue(input.Priority) != "50" {
		t.Fatalf("priority mismatch: %q", fastly.ToValue(input.Priority))
	}
	if fastly.ToValue(input.Type) != fastly.SnippetTypeDeliver {
		t.Fatalf("type mismatch: %q", fastly.ToValue(input.Type))
	}
}

func TestValidateConfig(t *testing.T) {
	content := types.StringValue(`set resp.http.X-Test = "true";`)

	tests := []struct {
		name    string
		items   []NestedModel
		wantErr bool
	}{
		{
			name: "empty list is valid",
		},
		{
			name: "single known snippet is valid",
			items: []NestedModel{
				{Name: types.StringValue("one"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
			},
		},
		{
			name: "duplicate known names are invalid",
			items: []NestedModel{
				{Name: types.StringValue("one"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
				{Name: types.StringValue("one"), Type: types.StringValue("deliver"), Priority: types.Int64Value(100), Content: content},
			},
			wantErr: true,
		},
		{
			name: "blank known name is invalid",
			items: []NestedModel{
				{Name: types.StringValue(" "), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
			},
			wantErr: true,
		},
		{
			name: "unknown name is deferred",
			items: []NestedModel{
				{Name: types.StringUnknown(), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.items)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	content := types.StringValue(`set resp.http.X-Test = "true";`)

	tests := []struct {
		name    string
		items   []NestedModel
		wantErr bool
	}{
		{
			name: "single snippet is valid",
			items: []NestedModel{
				{Name: types.StringValue("one"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
			},
		},
		{
			name: "duplicate names are invalid",
			items: []NestedModel{
				{Name: types.StringValue("one"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
				{Name: types.StringValue("one"), Type: types.StringValue("deliver"), Priority: types.Int64Value(100), Content: content},
			},
			wantErr: true,
		},
		{
			name: "invalid type is invalid",
			items: []NestedModel{
				{Name: types.StringValue("one"), Type: types.StringValue("invalid"), Priority: types.Int64Value(100), Content: content},
			},
			wantErr: true,
		},
		{
			name: "blank name is invalid",
			items: []NestedModel{
				{Name: types.StringValue(" "), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: content},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.items)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}

func TestContentEqual(t *testing.T) {
	if !ContentEqual("set resp.http.X = \"true\";", "set resp.http.X = \"true\";\n") {
		t.Fatal("expected trailing newline to be ignored")
	}

	if ContentEqual("set resp.http.X = \"true\";", "set resp.http.X = \"false\";") {
		t.Fatal("expected different content to be unequal")
	}
}

func TestMatchOrderPreservePlanContent(t *testing.T) {
	remote := []NestedModel{
		{Name: types.StringValue("b"), Type: types.StringValue("deliver"), Priority: types.Int64Value(100), Content: types.StringValue("remote b\n")},
		{Name: types.StringValue("a"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: types.StringValue("remote a\n")},
	}
	plan := []NestedModel{
		{Name: types.StringValue("a"), Type: types.StringValue("recv"), Priority: types.Int64Value(100), Content: types.StringValue("planned a")},
		{Name: types.StringValue("b"), Type: types.StringValue("deliver"), Priority: types.Int64Value(100), Content: types.StringValue("planned b")},
	}

	got := MatchOrderPreservePlanContent(remote, plan)
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].Name.ValueString() != "a" || got[0].Content.ValueString() != "planned a" {
		t.Fatalf("unexpected first item: %#v", got[0])
	}
	if got[1].Name.ValueString() != "b" || got[1].Content.ValueString() != "planned b" {
		t.Fatalf("unexpected second item: %#v", got[1])
	}
}
