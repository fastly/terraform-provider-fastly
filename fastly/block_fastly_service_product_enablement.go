package fastly

import (
	"context"
	"log"

	gofastly "github.com/fastly/go-fastly/v7/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProductEnablementServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ProductEnablementServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceProductEnablement returns a new resource.
func NewServiceProductEnablement(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ProductEnablementServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "product_enablement",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ProductEnablementServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ProductEnablementServiceAttributeHandler) GetSchema() *schema.Schema {
	blockAttributes := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Internal property used to calculate plan diff",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		blockAttributes["fanout"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Fanout support",
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["brotli_compression"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Brotli Compression support",
		}
		blockAttributes["domain_inspector"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Domain Inspector support",
		}
		blockAttributes["image_optimizer"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Image Optimizer support",
		}
		blockAttributes["origin_inspector"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Origin Inspector support",
		}
	}

	// websockets is supported for both Compute (wasm) and Deliver (vcl) services.
	blockAttributes["websockets"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Enable WebSockets support",
	}

	// NOTE: Min/MaxItems: 1 (to enforce only one product_enablement per service).
	// lintignore:S018
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: blockAttributes,
		},
	}
}

// Create creates the resource.
func (h *ProductEnablementServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	log.Printf("[DEBUG] Service ID: %s\n", serviceID)

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		if resource["fanout"].(bool) {
			log.Println("[DEBUG] fanout set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductFanout,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if resource["brotli_compression"].(bool) {
			log.Println("[DEBUG] brotli_compression set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductBrotliCompression,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}

		if resource["domain_inspector"].(bool) {
			log.Println("[DEBUG] domain_inspector set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductDomainInspector,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}

		if resource["image_optimizer"].(bool) {
			log.Println("[DEBUG] image_optimizer set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductImageOptimizer,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}

		if resource["origin_inspector"].(bool) {
			log.Println("[DEBUG] origin_inspector set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductOriginInspector,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}
	}

	if resource["websockets"].(bool) {
		log.Println("[DEBUG] websockets set")
		_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductWebSockets,
			ServiceID: serviceID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// Read refreshes the resource.
func (h *ProductEnablementServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.Key()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		log.Printf("[DEBUG] Refreshing Product Enablement Configuration for (%s)", d.Id())

		// The API returns a 400 if a product is not enabled.
		// The API client returns an error if a non-2xx is returned from the API.

		var (
			enabled bool
			err     error
		)

		// NOTE: We must set name to a unique value for a plan diff to be calculated.
		//
		// This value can be static because (like with the 'package' block) there can
		// only ever be one 'product_enablement' block per service resource.
		// This is ensured via the schema where we set MinItems/MaxItems to 1.
		//
		// Unlike the 'package' block we use a structure copied from 'backend'.
		// This is done so we can benefit from the 'modified' map data passed to the
		// 'update' CRUD method.
		result := map[string]any{
			"name": "product_enablement",
		}

		if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
			enabled = false
			_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductFanout,
				ServiceID: d.Id(),
			})
			if err == nil {
				enabled = true
			}
			result["fanout"] = enabled
		}

		if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
			enabled = false
			_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductBrotliCompression,
				ServiceID: d.Id(),
			})
			if err == nil {
				enabled = true
			}
			result["brotli_compression"] = enabled

			enabled = false
			_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductDomainInspector,
				ServiceID: d.Id(),
			})
			if err == nil {
				enabled = true
			}
			result["domain_inspector"] = enabled

			enabled = false
			_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductImageOptimizer,
				ServiceID: d.Id(),
			})
			if err == nil {
				enabled = true
			}
			result["image_optimizer"] = enabled

			enabled = false
			_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductOriginInspector,
				ServiceID: d.Id(),
			})
			if err == nil {
				enabled = true
			}
			result["origin_inspector"] = enabled
		}

		enabled = false
		_, err = conn.GetProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductWebSockets,
			ServiceID: d.Id(),
		})
		if err == nil {
			enabled = true
		}
		result["websockets"] = enabled

		results := []map[string]any{result}

		if err := d.Set(h.Key(), results); err != nil {
			log.Printf("[WARN] Error setting Product Enablement for (%s): %s", d.Id(), err)
			return err
		}
	}

	return nil
}

// Update updates the resource.
func (h *ProductEnablementServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, resource, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()

	log.Printf("[DEBUG] modified: %+v\n", modified)

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		if v, ok := modified["fanout"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] fanout set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductFanout,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] fanout not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductFanout,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if v, ok := modified["brotli_compression"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] brotli_compression set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductBrotliCompression,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] brotli_compression not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductBrotliCompression,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			}
		}

		if v, ok := modified["domain_inspector"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] domain_inspector set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductDomainInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] domain_inspector not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductDomainInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			}
		}

		if v, ok := modified["image_optimizer"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] image_optimizer set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductImageOptimizer,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] image_optimizer not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductImageOptimizer,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			}
		}

		if v, ok := modified["origin_inspector"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] origin_inspector set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductOriginInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] origin_inspector not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductOriginInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	if v, ok := modified["websockets"]; ok {
		if v.(bool) {
			log.Println("[DEBUG] websockets set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductWebSockets,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		} else {
			log.Println("[DEBUG] websockets not set")
			err := conn.DisableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductWebSockets,
				ServiceID: serviceID,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete deletes the resource.
//
// FIXME: This implementation causes unnecessary API calls.
// We should check if the specific 'products' have been disabled.
func (h *ProductEnablementServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		err := conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductFanout,
			ServiceID: d.Id(),
		})
		if err != nil {
			return err
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		err := conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductBrotliCompression,
			ServiceID: d.Id(),
		})
		if err != nil {
			return err
		}

		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductDomainInspector,
			ServiceID: d.Id(),
		})
		if err != nil {
			return err
		}

		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductImageOptimizer,
			ServiceID: d.Id(),
		})
		if err != nil {
			return err
		}

		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductOriginInspector,
			ServiceID: d.Id(),
		})
		if err != nil {
			return err
		}
	}

	err := conn.DisableProduct(&gofastly.ProductEnablementInput{
		ProductID: gofastly.ProductWebSockets,
		ServiceID: d.Id(),
	})
	if err != nil {
		return err
	}

	return nil
}
