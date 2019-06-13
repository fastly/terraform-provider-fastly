package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("FASTLY_API_KEY", nil),
				Description: "Fastly API Key from https://app.fastly.com/#account",
			},
			"base_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("FASTLY_API_URL", gofastly.DefaultEndpoint),
				Description: "Fastly API URL",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"fastly_ip_ranges": dataSourceFastlyIPRanges(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"fastly_service_v1":                  resourceServiceV1(),
			"fastly_service_dictionary_items_v1": resourceServiceDictionaryItemsV1(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ApiKey:  d.Get("api_key").(string),
		BaseURL: d.Get("base_url").(string),
	}
	return config.Client()
}
