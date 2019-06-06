package fastly

import (
	"fmt"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/httpclient"
	"github.com/terraform-providers/terraform-provider-fastly/version"
)

const TerraformProviderProductUserAgent = "terraform-provider-fastly"

type Config struct {
	ApiKey  string
	BaseURL string
}

type FastlyClient struct {
	conn *gofastly.Client
}

func (c *Config) Client() (interface{}, error) {
	var client FastlyClient

	if c.ApiKey == "" {
		return nil, fmt.Errorf("[Err] No API key for Fastly")
	}

	composedUserAgent := fmt.Sprintf("%s %s/%s", httpclient.UserAgentString(), TerraformProviderProductUserAgent, version.ProviderVersion)

	gofastly.UserAgent = composedUserAgent

	fastlyClient, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)
	if err != nil {
		return nil, err
	}

	fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", fastlyClient.HTTPClient.Transport)

	client.conn = fastlyClient
	return &client, nil
}
