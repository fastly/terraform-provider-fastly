package client

import (
	"fmt"
	"net/http"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/fastly/terraform-provider-fastly/internal/service"
)

type Data struct {
	Client             *fastly.Client
	VersionChecker     *service.VersionChecker
	ServiceTypeChecker *service.ServiceTypeChecker
}

func NewData(client *fastly.Client, userAgentPrefix string) *Data {
	baseHTTPClient := client.HTTPClient
	if baseHTTPClient == nil {
		baseHTTPClient = http.DefaultClient
	}

	baseTransport := baseHTTPClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	wrapped := *client
	wrapped.HTTPClient = &http.Client{
		Transport: &userAgentTransport{
			base:   baseTransport,
			prefix: userAgentPrefix,
		},
		Timeout: baseHTTPClient.Timeout,
	}

	return &Data{
		Client:             &wrapped,
		VersionChecker:     service.NewVersionChecker(&wrapped),
		ServiceTypeChecker: service.NewServiceTypeChecker(&wrapped),
	}
}

type userAgentTransport struct {
	base   http.RoundTripper
	prefix string
	suffix string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		ua = fastly.UserAgent
	}
	if t.prefix != "" {
		ua = t.prefix + " " + ua
	}
	if t.suffix != "" {
		ua = ua + " " + t.suffix
	}
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", ua)
	return t.base.RoundTrip(req)
}

func (d *Data) AutoClient() *fastly.Client {
	base := d.Client

	baseHTTPClient := base.HTTPClient
	if baseHTTPClient == nil {
		baseHTTPClient = http.DefaultClient
	}

	baseTransport := baseHTTPClient.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}

	existingTransport, ok := baseTransport.(*userAgentTransport)
	prefix := ""
	if ok {
		prefix = existingTransport.prefix
		baseTransport = existingTransport.base
	}

	wrapped := *base
	wrapped.HTTPClient = &http.Client{
		Transport: &userAgentTransport{
			base:   baseTransport,
			prefix: prefix,
			suffix: "mode=auto",
		},
		Timeout: baseHTTPClient.Timeout,
	}

	return &wrapped
}

func FromProviderData(raw any) (*Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	if raw == nil {
		return nil, diags
	}

	data, ok := raw.(*Data)
	if !ok {
		diags.AddError(
			"Unexpected ProviderData type",
			fmt.Sprintf("Expected *client.Data, got: %T", raw),
		)
		return nil, diags
	}

	return data, diags
}
