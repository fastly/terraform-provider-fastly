package fastly

import "testing"

func TestShouldRedact(t *testing.T) {
	rt := redactingTransport{
		redactKeys: []string{"Fastly-Key", "Authorization"},
	}

	tests := []struct {
		header   string
		expected bool
	}{
		{"Fastly-Key", true},
		{"Authorization", true},
		{"Content-Type", false},
	}

	for _, test := range tests {
		result := rt.shouldRedact(test.header)
		if result != test.expected {
			t.Errorf("expected redaction for %s to be %v, got %v", test.header, test.expected, result)
		}
	}
}
