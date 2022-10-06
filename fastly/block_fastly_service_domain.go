package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v6/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DomainServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DomainServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceDomain returns a new resource.
func NewServiceDomain(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&DomainServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "domain",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *DomainServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *DomainServiceAttributeHandler) GetSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "A set of Domain names to serve as entry points for your Service",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The domain that this Service will respond to. It is important to note that changing this attribute will delete and recreate the resource.",
				},

				"comment": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "An optional comment about the Domain.",
				},
			},
		},
	}
}

// Create creates the resource.
func (h *DomainServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.CreateDomainInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	if v, ok := resource["comment"]; ok {
		opts.Comment = v.(string)
	}

	log.Printf("[DEBUG] Fastly Domain Addition opts: %#v", opts)
	_, err := conn.CreateDomain(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Read refreshes the resource.
func (h *DomainServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	resources := d.Get(h.GetKey()).(*schema.Set).List()

	if len(resources) > 0 || d.Get("imported").(bool) {
		// TODO: update go-fastly to support an ActiveVersion struct, which contains
		// domain and backend info in the response. Here we do 2 additional queries
		// to find out that info
		log.Printf("[DEBUG] Refreshing Domains for (%s)", d.Id())
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			ServiceID:      d.Id(),
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return fmt.Errorf("error looking up Domains for (%s), version (%v): %s", d.Id(), serviceVersion, err)
		}

		// Refresh Domains
		dl := flattenDomains(domainList)

		if err := d.Set(h.GetKey(), dl); err != nil {
			log.Printf("[WARN] Error setting Domains for (%s): %s", d.Id(), err)
		}
	}

	return nil
}

// Update updates the resource.
func (h *DomainServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.UpdateDomainInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	if v, ok := modified["comment"]; ok {
		opts.Comment = gofastly.String(v.(string))
	}

	log.Printf("[DEBUG] Update Domain Opts: %#v", opts)
	_, err := conn.UpdateDomain(&opts)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the resource.
func (h *DomainServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	opts := gofastly.DeleteDomainInput{
		ServiceID:      d.Id(),
		ServiceVersion: serviceVersion,
		Name:           resource["name"].(string),
	}

	log.Printf("[DEBUG] Fastly Domain removal opts: %#v", opts)
	err := conn.DeleteDomain(&opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func flattenDomains(list []*gofastly.Domain) []map[string]any {
	dl := make([]map[string]any, 0, len(list))

	for _, d := range list {
		dl = append(dl, map[string]any{
			"name":    d.Name,
			"comment": d.Comment,
		})
	}

	return dl
}
