package fastly

import (
	"fmt"
	"net"
	"net/http"
	"time"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"golang.org/x/net/http2"
)

type Config struct {
	ApiKey     string
	BaseURL    string
	UserAgent  string
	NoAuth     bool
	ForceHttp2 bool
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

	// NOTE: We're fixing two issues here.
	// 1 (critical). go-fastly uses cleanhttp module that would disable keepalive connection:
	// https://github.com/hashicorp/go-cleanhttp/blob/v0.5.2/cleanhttp.go#L14-L15
	// this consumes local ports (source ports) more than necessary that could impact
	// some of the clients under restricted NAT environments such as:
	// https://github.com/fastly/terraform-provider-fastly/issues/484
	// overriding it with the default (still non-shared transport) so we can enable keepalive
	//
	// 2 (minor). while http.Transport supports HTTP/2 by default, it does TLS-ALPN negotiation
	// in order to support HTTP/1.x fallback. This means each new client connection initiated
	// by each resource will start TLS handshake regardless of the existing connection pool status.
	// explicitly assigning http2.Transport so there will be just one TLS-ALPN negotiation happening
	// (across all Fastly provider resources) against the same api.fastly.com:443 destination.
	httpDefaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// NOTE: "force_http2" provider option is an experimental feature.
	// http2.Transport struct fields are largely different than http.Transport
	// so leave it to default values for now.
	http2DefaultTransport := &http2.Transport{}

	if c.ForceHttp2 {
		fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", http2DefaultTransport)
	} else {
		fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", httpDefaultTransport)
	}

	client.conn = fastlyClient
	return &client, nil
}
