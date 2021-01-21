package fastly

import (
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	provider := &schema.Provider{
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
			"fastly_ip_ranges":         dataSourceFastlyIPRanges(),
			"fastly_tls_configuration": dataSourceFastlyTLSConfiguration(),
			"fastly_waf_rules":         dataSourceFastlyWAFRules(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"fastly_service_v1":                         resourceServiceV1(),
			"fastly_service_compute":                    resourceServiceComputeV1(),
			"fastly_service_acl_entries_v1":             resourceServiceAclEntriesV1(),
			"fastly_service_dictionary_items_v1":        resourceServiceDictionaryItemsV1(),
			"fastly_service_dynamic_snippet_content_v1": resourceServiceDynamicSnippetContentV1(),
			"fastly_user_v1":                            resourceUserV1(),
			"fastly_service_waf_configuration":          resourceServiceWAFConfigurationV1(),
			"fastly_tls_certificate":                    resourceTLSCertificate(),
			"fastly_tls_private_key":                    resourceTLSPrivateKey(),
		},
	}

	provider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return provider
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	config := Config{
		ApiKey:           d.Get("api_key").(string),
		BaseURL:          d.Get("base_url").(string),
		terraformVersion: terraformVersion,
	}
	return config.Client()
}
