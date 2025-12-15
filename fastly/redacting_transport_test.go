package fastly

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestRedactingTransport_ResponseBodyLogging(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		expectLogged  bool
		expectedInLog string
	}{
		{
			name:         "200 OK does not log response body",
			statusCode:   http.StatusOK,
			expectLogged: false,
		},
		{
			name:          "400 Bad Request logs response body",
			statusCode:    http.StatusBadRequest,
			expectLogged:  true,
			expectedInLog: "error: bad request",
		},
		{
			name:          "500 Internal Server Error logs response body",
			statusCode:    http.StatusInternalServerError,
			expectLogged:  true,
			expectedInLog: "error: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			ctx := context.Background()

			// Simulated body content
			body := tt.expectedInLog
			if body == "" {
				body = "ok"
			}

			// Set up mock HTTP server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(body))
			}))
			defer ts.Close()

			// Create transport with injected logger
			rt := &redactingTransport{
				ctx:        ctx,
				name:       "Fastly",
				underlying: http.DefaultTransport,
				redactKeys: []string{"Fastly-Key"},
				logf: func(_ context.Context, msg string) {
					logBuf.WriteString(msg + "\n")
				},
			}

			client := &http.Client{Transport: rt}
			resp, err := client.Get(ts.URL)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()
			_, _ = io.ReadAll(resp.Body)

			logOutput := logBuf.String()

			if tt.expectLogged {
				if !strings.Contains(logOutput, tt.expectedInLog) {
					t.Errorf("expected log to contain %q, but it did not.\nFull log output:\n%s", tt.expectedInLog, logOutput)
				}
			} else {
				if strings.Contains(logOutput, body) {
					t.Errorf("did not expect log to contain %q, but it was found.\nFull log output:\n%s", body, logOutput)
				}
			}
		})
	}
}
