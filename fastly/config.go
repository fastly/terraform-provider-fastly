package fastly

import (
	"fmt"

	"github.com/hashicorp/terraform/terraform"
	gofastly "github.com/sethvargo/go-fastly/fastly"
)

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

	gofastly.UserAgent = terraform.UserAgentString()

	fconn, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)

	if err != nil {
		return nil, err
	}

	client.conn = fconn
	return &client, nil
}
