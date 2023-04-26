package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SettingsServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type SettingsServiceAttributeHandler struct{}

// NewServiceSettings returns a new resource.
func NewServiceSettings() ServiceAttributeDefinition {
	return &SettingsServiceAttributeHandler{}
}

// Process creates or updates the attribute against the Fastly API.
func (h *SettingsServiceAttributeHandler) Process(_ context.Context, d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	// NOTE: DefaultTTL uses the same default value as provided by the Fastly API.
	opts := gofastly.UpdateSettingsInput{
		ServiceID:       d.Id(),
		ServiceVersion:  latestVersion,
		DefaultHost:     gofastly.String(d.Get("default_host").(string)),
		DefaultTTL:      uint(d.Get("default_ttl").(int)),
		StaleIfErrorTTL: gofastly.Uint(uint(d.Get("stale_if_error_ttl").(int))),
	}

	if attr, ok := d.GetOk("default_host"); ok {
		opts.DefaultHost = gofastly.String(attr.(string))
	}

	if attr, ok := d.GetOk("stale_if_error"); ok {
		opts.StaleIfError = gofastly.Bool(attr.(bool))
	}

	log.Printf("[DEBUG] Update Settings opts: %#v", opts)
	_, err := conn.UpdateSettings(&opts)

	if attr, ok := d.GetOk("http3"); ok {
		if attr.(bool) {
			// IMPORTANT: API will 400 when trying to enable HTTP3 when already on.
			//
			// So we first check the HTTP3 status.
			// The API returns a 404 if HTTP3 is not enabled.
			// The API client returns an error for non-2xx responses.
			// So if there is no error, then HTTP3 is enabled.
			if _, err = conn.GetHTTP3(&gofastly.GetHTTP3Input{
				ServiceID:      d.Id(),
				ServiceVersion: latestVersion,
			}); err != nil {
				_, err = conn.EnableHTTP3(&gofastly.EnableHTTP3Input{
					FeatureRevision: gofastly.Int(1),
					ServiceID:       d.Id(),
					ServiceVersion:  latestVersion,
				})
			}
		}
	} else {
		err = conn.DisableHTTP3(&gofastly.DisableHTTP3Input{
			ServiceID:      d.Id(),
			ServiceVersion: latestVersion,
		})
	}

	return err
}

func (h *SettingsServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	settingsOpts := gofastly.GetSettingsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	}

	settings, err := conn.GetSettings(&settingsOpts)
	if err != nil {
		return fmt.Errorf("error looking up Version settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	d.Set("default_host", settings.DefaultHost)
	d.Set("default_ttl", int(settings.DefaultTTL))
	d.Set("http3", false)
	d.Set("stale_if_error", bool(settings.StaleIfError))
	d.Set("stale_if_error_ttl", int(settings.StaleIfErrorTTL))

	// The API returns a 404 if HTTP3 is not enabled.
	// The API client returns an error for non-2xx responses.
	// So if there is no error, then HTTP3 is enabled.
	if _, err = conn.GetHTTP3(&gofastly.GetHTTP3Input{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	}); err == nil {
		d.Set("http3", true)
	}

	return nil
}

// HasChange returns whether the state of the attribute has changed against Terraform stored state.
func (h *SettingsServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChanges("default_ttl", "default_host", "http3", "stale_if_error", "stale_if_error_ttl")
}

// MustProcess returns whether we must process the resource
//
// If the requested default_ttl is 0, and this is the first
// version being created, HasChange will return false, but we need
// to set it anyway, so ensure we update the settings in that
// case.
func (h *SettingsServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d) || (d.Get("default_ttl") == 0 && initialVersion)
}

// Register add the attribute to the resource schema.
func (h *SettingsServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema["default_ttl"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     3600,
		Description: "The default Time-to-live (TTL) for requests",
	}
	s.Schema["default_host"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "The default hostname",
	}
	s.Schema["http3"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enables support for the HTTP/3 (QUIC) protocol",
	}
	s.Schema["stale_if_error"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enables serving a stale object if there is an error",
	}
	s.Schema["stale_if_error_ttl"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     43200,
		Description: "The default time-to-live (TTL) for serving the stale object for the version",
	}
	return nil
}
