package dynamicsnippet

import (
	"testing"

	regularsnippet "github.com/fastly/terraform-provider-fastly/internal/resources/snippet"

	fastly "github.com/fastly/go-fastly/v17/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateInput(t *testing.T) {
	model := NestedModel{
		Name:     types.StringValue("dynamic_recv"),
		Type:     types.StringValue("recv"),
		Priority: types.Int64Value(50),
	}

	input := BuildCreateInput("service-id", 1, model)

	if input.ServiceID != "service-id" {
		t.Fatalf("ServiceID = %q, want service-id", input.ServiceID)
	}
	if input.ServiceVersion != 1 {
		t.Fatalf("ServiceVersion = %d, want 1", input.ServiceVersion)
	}
	if fastly.ToValue(input.Name) != "dynamic_recv" {
		t.Fatalf("Name = %q, want dynamic_recv", fastly.ToValue(input.Name))
	}
	if fastly.ToValue(input.Dynamic) != 1 {
		t.Fatalf("Dynamic = %d, want 1", fastly.ToValue(input.Dynamic))
	}
	if fastly.ToValue(input.Priority) != "50" {
		t.Fatalf("Priority = %q, want 50", fastly.ToValue(input.Priority))
	}
}

func TestFlattenToNestedModelRejectsRegularSnippet(t *testing.T) {
	dynamic := 0
	name := "regular_recv"

	_, err := FlattenToNestedModel(&fastly.Snippet{
		Name:    &name,
		Dynamic: &dynamic,
	})
	if err == nil {
		t.Fatal("expected regular snippet to be rejected")
	}
}

func TestParsePriority(t *testing.T) {
	value := "25"
	got, err := parsePriority(&value)
	if err != nil {
		t.Fatalf("parsePriority returned error: %s", err)
	}
	if got != 25 {
		t.Fatalf("parsePriority = %d, want 25", got)
	}

	empty := ""
	got, err = parsePriority(&empty)
	if err != nil {
		t.Fatalf("parsePriority empty returned error: %s", err)
	}
	if got != DefaultPriority {
		t.Fatalf("parsePriority empty = %d, want %d", got, DefaultPriority)
	}

	invalid := "invalid"
	if _, err := parsePriority(&invalid); err == nil {
		t.Fatal("expected invalid priority to return an error")
	}
}

func TestValidateNoNameConflicts(t *testing.T) {
	dynamic := []NestedModel{
		{Name: types.StringValue("shared")},
	}
	regular := []regularsnippet.NestedModel{
		{Name: types.StringValue("shared")},
	}

	if err := ValidateNoNameConflicts(dynamic, regular); err == nil {
		t.Fatal("expected shared regular and dynamic snippet name to return an error")
	}
}
