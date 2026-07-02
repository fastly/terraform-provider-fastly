package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"
)

func TestUserAgentTransport(t *testing.T) {
	tests := []struct {
		name           string
		prefix         string
		suffix         string
		expectedSuffix string
	}{
		{
			name:           "prefix only",
			prefix:         "terraform-provider-fastly/1.2.3",
			suffix:         "",
			expectedSuffix: "terraform-provider-fastly/1.2.3 " + fastly.UserAgent,
		},
		{
			name:           "prefix and suffix",
			prefix:         "terraform-provider-fastly/1.2.3",
			suffix:         "mode=auto",
			expectedSuffix: "terraform-provider-fastly/1.2.3 " + fastly.UserAgent + " mode=auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUA string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedUA = r.Header.Get("User-Agent")
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			transport := &userAgentTransport{
				base:   http.DefaultTransport,
				prefix: tt.prefix,
				suffix: tt.suffix,
			}

			client := &http.Client{Transport: transport}
			req, _ := http.NewRequest("GET", server.URL, nil)
			_, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if capturedUA != tt.expectedSuffix {
				t.Errorf("Expected User-Agent %q, got %q", tt.expectedSuffix, capturedUA)
			}
		})
	}
}
