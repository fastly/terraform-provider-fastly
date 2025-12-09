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
}

func (rt *redactingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tflog.Debug(rt.ctx, rt.name+" --> "+req.Method+" "+req.URL.String())

	for key, values := range req.Header {
		value := strings.Join(values, ", ")
		if rt.shouldRedact(key) {
			value = "[REDACTED]"
		}
		tflog.Debug(rt.ctx, rt.name+" Header: "+key+": "+value)
	}

	if req.Body != nil && req.ContentLength > 0 {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset body
		tflog.Debug(rt.ctx, rt.name+" Body: "+string(body))
	}

	resp, err := rt.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	tflog.Debug(rt.ctx, rt.name+" <-- "+resp.Status)

	for key, values := range resp.Header {
		tflog.Debug(rt.ctx, rt.name+" Response Header: "+key+": "+strings.Join(values, ", "))
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
