package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

type SettingsServiceAttributeHandler struct {
}

func NewServiceSettings() ServiceAttributeDefinition {
	return &SettingsServiceAttributeHandler{}
}

func (h *SettingsServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateSettingsInput{
		Service: d.Id(),
		Version: latestVersion,
		// default_ttl has the same default value of 3600 that is provided by
		// the Fastly API, so it's safe to include here
		DefaultTTL: uint(d.Get("default_ttl").(int)),
	}

	if attr, ok := d.GetOk("default_host"); ok {
		opts.DefaultHost = attr.(string)
	}

	log.Printf("[DEBUG] Update Settings opts: %#v", opts)
	_, err := conn.UpdateSettings(&opts)

	return err
}

func (h *SettingsServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	settingsOpts := gofastly.GetSettingsInput{
		Service: d.Id(),
		Version: s.ActiveVersion.Number,
	}
	if settings, err := conn.GetSettings(&settingsOpts); err == nil {
		d.Set("default_host", settings.DefaultHost)
		d.Set("default_ttl", settings.DefaultTTL)
	} else {
		return fmt.Errorf("[ERR] Error looking up Version settings for (%s), version (%v): %s", d.Id(), s.ActiveVersion.Number, err)
	}
	return nil
}

func (h *SettingsServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChange("default_ttl") || d.HasChange("default_host")
}

// If the requested default_ttl is 0, and this is the first
// version being created, HasChange will return false, but we need
// to set it anyway, so ensure we update the settings in that
// case.
func (h *SettingsServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return d.HasChange("default_host") || d.HasChange("default_ttl") || (d.Get("default_ttl") == 0 && initialVersion)
}

func (h *SettingsServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema["default_ttl"] = &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     3600,
		Description: "The default Time-to-live (TTL) for the version",
	}
	s.Schema["default_host"] = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "The default hostname for the version",
	}
	return nil
}
