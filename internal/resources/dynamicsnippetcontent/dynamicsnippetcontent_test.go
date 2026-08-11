package dynamicsnippetcontent

import "testing"

func TestID(t *testing.T) {
	got := ID("service-id", "snippet-id")
	want := "service-id/snippet-id"
	if got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestParseImportID(t *testing.T) {
	serviceID, snippetID, err := parseImportID("service-id/snippet-id")
	if err != nil {
		t.Fatalf("parseImportID returned error: %s", err)
	}
	if serviceID != "service-id" {
		t.Fatalf("serviceID = %q, want service-id", serviceID)
	}
	if snippetID != "snippet-id" {
		t.Fatalf("snippetID = %q, want snippet-id", snippetID)
	}

	if _, _, err := parseImportID("service-id"); err == nil {
		t.Fatal("expected missing snippet_id to return an error")
	}
}
