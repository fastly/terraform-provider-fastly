package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

type Config struct {
	ApiKey    string
	BaseURL   string
	UserAgent string
}

type FastlyClient struct {
	conn *gofastly.Client
}

func (c *Config) Client() (*FastlyClient, diag.Diagnostics) {
	var client FastlyClient

	if c.ApiKey == "" {
		return nil, diag.FromErr(fmt.Errorf("[Err] No API key for Fastly"))
	}

	gofastly.UserAgent = c.UserAgent

	fastlyClient, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", fastlyClient.HTTPClient.Transport)

	client.conn = fastlyClient
	return &client, nil
}
