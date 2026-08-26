package fastly

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type redactingTransport struct {
	ctx        context.Context
	name       string
	underlying http.RoundTripper
	redactKeys []string

	// Optional: allows overriding logging behavior (e.g., in tests)
	logf func(ctx context.Context, msg string)
}

func (rt *redactingTransport) log(ctx context.Context, msg string) {
	if rt.logf != nil {
		rt.logf(ctx, msg)
	} else {
		tflog.Debug(ctx, msg)
	}
}

func (rt *redactingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.log(rt.ctx, rt.name+" --> "+req.Method+" "+req.URL.String())

	for key, values := range req.Header {
		value := strings.Join(values, ", ")
		if rt.shouldRedact(key) {
			value = "[REDACTED]"
		}
		rt.log(rt.ctx, rt.name+" Header: "+key+": "+value)
	}

	if req.Body != nil && req.ContentLength > 0 {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body
		rt.log(rt.ctx, rt.name+" Body: "+string(body))
	}

	resp, err := rt.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	rt.log(rt.ctx, rt.name+" <-- "+resp.Status)

	for key, values := range resp.Header {
		rt.log(rt.ctx, rt.name+" Response Header: "+key+": "+strings.Join(values, ", "))
	}

	if resp.Body != nil && resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body

		preview := string(body)
		if len(preview) > 300 {
			preview = preview[:300] + "...[truncated]"
		}

		rt.log(rt.ctx, rt.name+" Response Body: "+preview)
	}

	return resp, nil
}

func (rt *redactingTransport) shouldRedact(header string) bool {
	header = strings.ToLower(header)
	for _, k := range rt.redactKeys {
		if strings.ToLower(k) == header {
			return true
		}
	}
	return false
}
