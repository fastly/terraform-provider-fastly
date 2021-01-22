package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceFastlyTLSConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSConfigurationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Custom name of the TLS configuration",
				Optional:    true,
				Computed:    true,
			},
			"tls_protocols": {
				Type:        schema.TypeSet,
				Description: "TLS protocols available on the TLS configuration",
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"http_protocols": {
				Type:        schema.TypeSet,
				Description: "HTTP protocols available on the TLS configuration",
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tls_service": {
				Type:         schema.TypeString,
				Description:  fmt.Sprintf("Whether the configuration should support the `%s` or `%s` TLS service", tlsPlatformService, tlsCustomService),
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{tlsPlatformService, tlsCustomService}, false),
			},
			"default": {
				Type:        schema.TypeBool,
				Description: "Signifies whether Fastly will use this configuration as a default when creating a new TLS activation",
				Optional:    true,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Time-stamp (GMT) when the configuration was created",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Time-stamp (GMT) when the configuration was last updated",
				Computed:    true,
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

	var configurations []*fastly.CustomTLSConfiguration
	cursor := 0
	for {
		list, err := conn.ListCustomTLSConfigurations(&fastly.ListCustomTLSConfigurationsInput{
			PageNumber: cursor,
		})
		if err != nil {
			return err
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

	if len(configurations) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(configurations) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please change try a more specific search criteria and try again.")
	}

	configuration := configurations[0]

	tlsService := tlsCustomService
	if configuration.Bulk {
		tlsService = tlsPlatformService
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
