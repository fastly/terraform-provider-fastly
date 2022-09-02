package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
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

	return err
}

func (h *SettingsServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	settingsOpts := gofastly.GetSettingsInput{
		ServiceID:      d.Id(),
		ServiceVersion: s.ActiveVersion.Number,
	}

	settings, err := conn.GetSettings(&settingsOpts)
	if err != nil {
		return fmt.Errorf("[ERR] Error looking up Version settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}

	d.Set("default_host", settings.DefaultHost)
	d.Set("default_ttl", int(settings.DefaultTTL))
	d.Set("stale_if_error", bool(settings.StaleIfError))
	d.Set("stale_if_error_ttl", int(settings.StaleIfErrorTTL))

	return nil
}

// HasChange returns whether the state of the attribute has changed against Terraform stored state.
func (h *SettingsServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChanges("default_ttl", "default_host", "stale_if_error", "stale_if_error_ttl")
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
