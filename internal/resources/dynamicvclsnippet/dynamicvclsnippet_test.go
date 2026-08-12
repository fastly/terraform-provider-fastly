package dynamicvclsnippet

import "testing"

func TestID(t *testing.T) {
	got := ID("service123", 3, "block_scrapers")
	want := "service123-3-block_scrapers"

	if got != want {
		t.Fatalf("ID() = %q, want %q", got, want)
	}
}
