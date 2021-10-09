package fastly

import (
	"fmt"

	gofastly "github.com/fastly/go-fastly/v5/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/net/http2"
)

type Config struct {
	ApiKey    string
	BaseURL   string
	UserAgent string
	NoAuth    bool
}

type FastlyClient struct {
	conn *gofastly.Client
}

func (c *Config) Client() (*FastlyClient, diag.Diagnostics) {
	var client FastlyClient

	if !c.NoAuth && c.ApiKey == "" {
		return nil, diag.FromErr(fmt.Errorf("[Err] No API key for Fastly"))
	}

	gofastly.UserAgent = c.UserAgent

	fastlyClient, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	// explicitly assigning http2.Transport so there will be just one TLS-ALPN negotiation happens
	// amoung all Fastly provider resources against the same api.fastly.com:443 destination.
	// otherwise, each resource would consume a different source port due to TLS handshake.
	// see also: https://github.com/fastly/terraform-provider-fastly/issues/484
	fastlyClient.HTTPClient.Transport = &http2.Transport{}
	fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", fastlyClient.HTTPClient.Transport)

	client.conn = fastlyClient
	return &client, nil
}
