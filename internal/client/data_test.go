package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v17/fastly"
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

func TestRetryTransport_MethodRouting(t *testing.T) {
	const retryMax = 2

	tests := []struct {
		name         string
		method       string
		statusCode   int
		wantAttempts int
	}{
		{name: "GET retried on 503", method: http.MethodGet, statusCode: http.StatusServiceUnavailable, wantAttempts: retryMax + 1},
		{name: "HEAD retried on 503", method: http.MethodHead, statusCode: http.StatusServiceUnavailable, wantAttempts: retryMax + 1},
		{name: "PUT retried on 503", method: http.MethodPut, statusCode: http.StatusServiceUnavailable, wantAttempts: retryMax + 1},
		{name: "DELETE retried on 503", method: http.MethodDelete, statusCode: http.StatusServiceUnavailable, wantAttempts: retryMax + 1},
		{name: "POST not retried on 503", method: http.MethodPost, statusCode: http.StatusServiceUnavailable, wantAttempts: 1},
		{name: "PATCH not retried on 503", method: http.MethodPatch, statusCode: http.StatusServiceUnavailable, wantAttempts: 1},
		{name: "GET not retried on 429", method: http.MethodGet, statusCode: http.StatusTooManyRequests, wantAttempts: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts++
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			transport := newRetryTransport(http.DefaultTransport).(*retryTransport)
			transport.retryClient.RetryMax = retryMax
			transport.retryClient.RetryWaitMin = time.Millisecond
			transport.retryClient.RetryWaitMax = time.Millisecond

			client := &http.Client{Transport: transport}
			req, err := http.NewRequest(tt.method, server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}

			if attempts != tt.wantAttempts {
				t.Errorf("Expected %d attempts, got %d", tt.wantAttempts, attempts)
			}
		})
	}
}
