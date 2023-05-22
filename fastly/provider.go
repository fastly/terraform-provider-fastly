package fastly

import (
	"context"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/fastly/terraform-provider-fastly/version"
)

// TerraformProviderProductUserAgent is included in the User-Agent header for
// any API requests made by the provider.
const TerraformProviderProductUserAgent = "terraform-provider-fastly"

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
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
			"force_http2": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set this to `true` to disable HTTP/1.x fallback mechanism that the underlying Go library will attempt upon connection to `api.fastly.com:443` by default. This may slightly improve the provider's performance and reduce unnecessary TLS handshakes. Default: `false`",
			},
			"no_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set to `true` if your configuration only consumes data sources that do not require authentication, such as `fastly_ip_ranges`",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"fastly_datacenters":                  dataSourceFastlyDatacenters(),
			"fastly_dictionaries":                 dataSourceFastlyDictionaries(),
			"fastly_ip_ranges":                    dataSourceFastlyIPRanges(),
			"fastly_package_hash":                 dataSourceFastlyPackageHash(),
			"fastly_services":                     dataSourceFastlyServices(),
			"fastly_tls_activation":               dataSourceFastlyTLSActivation(),
			"fastly_tls_activation_ids":           dataSourceFastlyTLSActivationIds(),
			"fastly_tls_certificate":              dataSourceFastlyTLSCertificate(),
			"fastly_tls_certificate_ids":          dataSourceFastlyTLSCertificateIDs(),
			"fastly_tls_configuration":            dataSourceFastlyTLSConfiguration(),
			"fastly_tls_configuration_ids":        dataSourceFastlyTLSConfigurationIDs(),
			"fastly_tls_domain":                   dataSourceFastlyTLSDomain(),
			"fastly_tls_platform_certificate":     dataSourceFastlyTLSPlatformCertificate(),
			"fastly_tls_platform_certificate_ids": dataSourceFastlyTLSPlatformCertificateIDs(),
			"fastly_tls_private_key":              dataSourceFastlyTLSPrivateKey(),
			"fastly_tls_private_key_ids":          dataSourceFastlyTLSPrivateKeyIDs(),
			"fastly_tls_subscription":             dataSourceFastlyTLSSubscription(),
			"fastly_tls_subscription_ids":         dataSourceFastlyTLSSubscriptionIDs(),
			"fastly_waf_rules":                    dataSourceFastlyWAFRules(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"fastly_service_acl_entries":             resourceServiceACLEntries(),
			"fastly_service_authorization":           resourceServiceAuthorization(),
			"fastly_service_compute":                 resourceServiceCompute(),
			"fastly_service_dictionary_items":        resourceServiceDictionaryItems(),
			"fastly_service_dynamic_snippet_content": resourceServiceDynamicSnippetContent(),
			"fastly_service_vcl":                     resourceServiceVCL(),
			"fastly_service_waf_configuration":       resourceServiceWAFConfiguration(),
			"fastly_tls_activation":                  resourceFastlyTLSActivation(),
			"fastly_tls_certificate":                 resourceFastlyTLSCertificate(),
			"fastly_tls_platform_certificate":        resourceFastlyTLSPlatformCertificate(),
			"fastly_tls_private_key":                 resourceFastlyTLSPrivateKey(),
			"fastly_tls_subscription":                resourceFastlyTLSSubscription(),
			"fastly_tls_subscription_validation":     resourceFastlyTLSSubscriptionValidation(),
			"fastly_user":                            resourceUser(),
		},
	}

	provider.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		config := Config{
			APIKey:     d.Get("api_key").(string),
			BaseURL:    d.Get("base_url").(string),
			NoAuth:     d.Get("no_auth").(bool),
			ForceHTTP2: d.Get("force_http2").(bool),
			UserAgent:  provider.UserAgent(TerraformProviderProductUserAgent, version.ProviderVersion),
		}
		return config.Client()
	}

	return provider
}
