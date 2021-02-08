package fastly

import (
	"fmt"
	"os"
	"testing"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const testResourcePrefix = "tf-test"

var sweeperClients map[string]*fastly.Client

func TestMain(m *testing.M) {
	sweeperClients = make(map[string]*fastly.Client)
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*fastly.Client, error) {
	if client, ok := sweeperClients[region]; ok {
		return client, nil
	}

	if os.Getenv("FASTLY_API_KEY") == "" {
		return nil, fmt.Errorf("must provide environment variables 'FASTLY_API_KEY'")
	}

	url := fastly.DefaultEndpoint
	if v := os.Getenv("FASTLY_API_URL"); v != "" {
		url = v
	}
	c := Config{
		ApiKey:           os.Getenv("FASTLY_API_KEY"),
		BaseURL:          url,
		terraformVersion: "test-sweepers",
	}

	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	sweeperClients[region] = client.conn

	return client.conn, nil
}
