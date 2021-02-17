package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceFastlyTLSConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSConfigurationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Description:   "ID of the TLS configuration obtained from the Fastly API or another data source. Conflicts with all the other filters.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name", "tls_protocols", "http_protocols", "tls_service", "default"},
			},
			"name": {
				Type:          schema.TypeString,
				Description:   "Custom name of the TLS configuration.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"tls_protocols": {
				Type:          schema.TypeSet,
				Description:   "TLS protocols available on the TLS configuration.",
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"id"},
			},
			"http_protocols": {
				Type:          schema.TypeSet,
				Description:   "HTTP protocols available on the TLS configuration.",
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"id"},
			},
			"tls_service": {
				Type:          schema.TypeString,
				Description:   fmt.Sprintf("Whether the configuration should support the `%s` or `%s` TLS service.", tlsPlatformService, tlsCustomService),
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.StringInSlice([]string{tlsPlatformService, tlsCustomService}, false),
				ConflictsWith: []string{"id"},
			},
			"default": {
				Type:          schema.TypeBool,
				Description:   "Signifies whether Fastly will use this configuration as a default when creating a new TLS activation.",
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"id"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the configuration was created.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the configuration was last updated.",
				Computed:    true,
			},
			"dns_records": {
				Type:        schema.TypeSet,
				Description: "The available DNS addresses that can be used to enable TLS for a domain. DNS must be configured for a domain for TLS handshakes to succeed. If enabling TLS on an apex domain (e.g. `example.com`) you must create four A records (or four AAAA records for IPv6 support) using the displayed global A record's IP addresses with your DNS provider. For subdomains and wildcard domains (e.g. `www.example.com` or `*.example.com`) you will need to create a relevant CNAME record.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"record_type": {
							Type:        schema.TypeString,
							Description: "Type of DNS record to set, e.g. A, AAAA, or CNAME.",
							Computed:    true,
						},
						"record_value": {
							Type:        schema.TypeString,
							Description: "The IP address or hostname of the DNS record.",
							Computed:    true,
						},
						"region": {
							Type:        schema.TypeString,
							Description: "The regions that will be used to route traffic. Select DNS Records with a `global` region to route traffic to the most performant point of presence (POP) worldwide (global pricing will apply). Select DNS records with a `us-eu` region to exclusively land traffic on North American and European POPs.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

const (
	tlsPlatformService = "PLATFORM"
	tlsCustomService   = "CUSTOM"
)

func dataSourceFastlyTLSConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var configuration *fastly.CustomTLSConfiguration

	if v, ok := d.GetOk("id"); ok {
		config, err := conn.GetCustomTLSConfiguration(&fastly.GetCustomTLSConfigurationInput{
			ID: v.(string),
		})
		if err != nil {
			return err
		}

		configuration = config
	} else {
		filters := getTLSConfigurationFilters(d)

		configurations, err := listTLSConfigurations(conn, filters...)
		if err != nil {
			return err
		}

		if len(configurations) == 0 {
			return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
		}

		if len(configurations) > 1 {
			return fmt.Errorf("Your query returned more than one result. Please change try a more specific search criteria and try again.")
		}

		configuration = configurations[0]
	}

	return dataSourceFastlyTLSConfigurationSetAttributes(configuration, d)
}

func getTLSConfigurationFilters(d *schema.ResourceData) []func(*fastly.CustomTLSConfiguration) bool {
	var filters []func(*fastly.CustomTLSConfiguration) bool

	if v, ok := d.GetOk("name"); ok {
		filters = append(filters, func(c *fastly.CustomTLSConfiguration) bool {
			return c.Name == v.(string)
		})
	}
	if v, ok := d.GetOk("tls_protocols"); ok {
		filters = append(filters, func(c *fastly.CustomTLSConfiguration) bool {
			return containsSubSet(c.TLSProtocols, v.(*schema.Set).List())
		})
	}
	if v, ok := d.GetOk("http_protocols"); ok {
		filters = append(filters, func(c *fastly.CustomTLSConfiguration) bool {
			return containsSubSet(c.HTTPProtocols, v.(*schema.Set).List())
		})
	}
	if v, ok := d.GetOk("tls_service"); ok {
		service := v.(string)
		filters = append(filters, func(c *fastly.CustomTLSConfiguration) bool {
			if service == tlsPlatformService {
				return c.Bulk == true
			}
			return c.Bulk == false
		})
	}
	if v, ok := d.GetOk("default"); ok {
		filters = append(filters, func(c *fastly.CustomTLSConfiguration) bool {
			return c.Default == v.(bool)
		})
	}

	return filters
}

func listTLSConfigurations(conn *fastly.Client, filters ...func(*fastly.CustomTLSConfiguration) bool) ([]*fastly.CustomTLSConfiguration, error) {
	var configurations []*fastly.CustomTLSConfiguration
	cursor := 0
	for {
		list, err := conn.ListCustomTLSConfigurations(&fastly.ListCustomTLSConfigurationsInput{
			PageNumber: cursor,
			Include:    "dns_records",
		})
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		cursor += len(list)

		for _, configuration := range list {
			if filterTLSConfiguration(configuration, filters) {
				configurations = append(configurations, configuration)
			}
		}
	}
	return configurations, nil
}

func dataSourceFastlyTLSConfigurationSetAttributes(configuration *fastly.CustomTLSConfiguration, d *schema.ResourceData) error {
	tlsService := tlsCustomService
	if configuration.Bulk {
		tlsService = tlsPlatformService
	}

	var DNSRecords []map[string]string
	for _, record := range configuration.DNSRecords {
		DNSRecords = append(DNSRecords, map[string]string{
			"record_type":  record.RecordType,
			"record_value": record.ID,
			"region":       record.Region,
		})
	}

	d.SetId(configuration.ID)
	if err := d.Set("name", configuration.Name); err != nil {
		return err
	}
	if err := d.Set("tls_protocols", configuration.TLSProtocols); err != nil {
		return err
	}
	if err := d.Set("http_protocols", configuration.HTTPProtocols); err != nil {
		return err
	}
	if err := d.Set("tls_service", tlsService); err != nil {
		return err
	}
	if err := d.Set("default", configuration.Default); err != nil {
		return err
	}
	if err := d.Set("created_at", configuration.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", configuration.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("dns_records", DNSRecords); err != nil {
		return err
	}
	return nil
}

func filterTLSConfiguration(config *fastly.CustomTLSConfiguration, filters []func(*fastly.CustomTLSConfiguration) bool) bool {
	for _, f := range filters {
		if !f(config) {
			return false
		}
	}
	return true
}

func containsSubSet(set []string, subSet []interface{}) bool {
	for _, s := range subSet {
		if !contains(set, s) {
			return false
		}
	}
	return true
}

func contains(haystack []string, needle interface{}) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}

	return false
}
