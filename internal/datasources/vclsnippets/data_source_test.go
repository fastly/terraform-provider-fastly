package vclsnippets

import (
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"
)

func TestFlattenVCLSnippets(t *testing.T) {
	name := "recv_snippet"
	content := "set req.http.X-Test = \"regular\";"
	priority := "10"
	snippetType := fastly.SnippetType("recv")
	snippetID := "snippet-id"
	dynamic := 0

	setVal, ids, diags := flattenVCLSnippets([]*fastly.Snippet{
		{
			Name:      &name,
			Content:   &content,
			Priority:  &priority,
			Type:      &snippetType,
			SnippetID: &snippetID,
			Dynamic:   &dynamic,
		},
	})
	if diags.HasError() {
		t.Fatalf("flattenVCLSnippets returned diagnostics: %v", diags)
	}

	if len(ids) != 1 {
		t.Fatalf("ids length = %d, want 1", len(ids))
	}

	if setVal.IsNull() || setVal.IsUnknown() {
		t.Fatal("expected non-null, known set")
	}

	if len(setVal.Elements()) != 1 {
		t.Fatalf("set elements length = %d, want 1", len(setVal.Elements()))
	}
}

func TestFlattenVCLSnippetsInvalidPriority(t *testing.T) {
	priority := "invalid"

	_, _, diags := flattenVCLSnippets([]*fastly.Snippet{
		{
			Priority: &priority,
		},
	})
	if !diags.HasError() {
		t.Fatal("expected invalid priority to return diagnostics")
	}
}

func TestParsePriority(t *testing.T) {
	got, err := parsePriority(nil)
	if err != nil {
		t.Fatalf("parsePriority nil returned error: %s", err)
	}
	if got != defaultPriority {
		t.Fatalf("parsePriority nil = %d, want %d", got, defaultPriority)
	}

	value := "25"
	got, err = parsePriority(&value)
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
	if got != defaultPriority {
		t.Fatalf("parsePriority empty = %d, want %d", got, defaultPriority)
	}
}
