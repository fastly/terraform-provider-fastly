package fastly

import (
	"fmt"

	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/fastly/terraform-provider-fastly/version"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/httpclient"
)

const TerraformProviderProductUserAgent = "terraform-provider-fastly"

type Config struct {
	ApiKey  string
	BaseURL string

	terraformVersion string
}

type FastlyClient struct {
	conn *gofastly.Client
}

func (c *Config) Client() (*FastlyClient, error) {
	var client FastlyClient

	if c.ApiKey == "" {
		return nil, fmt.Errorf("[Err] No API key for Fastly")
	}

	tfUserAgent := httpclient.TerraformUserAgent(c.terraformVersion)
	providerUserAgent := fmt.Sprintf("%s/%s", TerraformProviderProductUserAgent, version.ProviderVersion)
	ua := fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent)
	gofastly.UserAgent = ua

	fastlyClient, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)
	if err != nil {
		return nil, err
	}

	fastlyClient.HTTPClient.Transport = logging.NewTransport("Fastly", fastlyClient.HTTPClient.Transport)

	client.conn = fastlyClient
	return &client, nil
}
