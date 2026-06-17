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

func NewData(client *fastly.Client) *Data {
	return &Data{
		Client:             client,
		VersionChecker:     service.NewVersionChecker(client),
		ServiceTypeChecker: service.NewServiceTypeChecker(client),
	}
}

type userAgentTransport struct {
	base   http.RoundTripper
	suffix string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		ua = fastly.UserAgent
	}
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", ua+" "+t.suffix)
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

	wrapped := *base
	wrapped.HTTPClient = &http.Client{
		Transport: &userAgentTransport{
			base:   baseTransport,
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
