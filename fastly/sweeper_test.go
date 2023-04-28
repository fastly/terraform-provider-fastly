package fastly

import (
	"fmt"
	"os"
	"testing"

	"github.com/fastly/terraform-provider-fastly/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testResourcePrefix = "tf-test"

var sweeperClients map[string]*fastly.Client

func TestMain(m *testing.M) {
	sweeperClients = make(map[string]*fastly.Client)
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*fastly.Client, diag.Diagnostics) {
	if client, ok := sweeperClients[region]; ok {
		return client, nil
	}

	url := fastly.DefaultEndpoint
	if v := os.Getenv("FASTLY_API_URL"); v != "" {
		url = v
	}
	c := Config{
		APIKey:  os.Getenv("FASTLY_API_KEY"),
		BaseURL: url,
		UserAgent: fmt.Sprintf(
			"HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s %s/%s",
			"test-sweepers",
			meta.SDKVersionString(),
			TerraformProviderProductUserAgent,
			version.ProviderVersion,
		),
	}

	client, diagnostics := c.Client()
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	sweeperClients[region] = client.conn

	return client.conn, nil
}
