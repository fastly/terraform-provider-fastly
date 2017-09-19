package fastly

import (
	"fmt"

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

	if c.BaseURL == "" {
		c.BaseURL = gofastly.DefaultEndpoint
	}

	fconn, err := gofastly.NewClientForEndpoint(c.ApiKey, c.BaseURL)
	if err != nil {
		return nil, err
	}

	client.conn = fconn
	return &client, nil
}
