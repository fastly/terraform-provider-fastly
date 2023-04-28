package fastly

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

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
			Description: "Used internally by the provider to identify modified settings",
		},
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		blockAttributes["fanout"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Fanout support",
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["brotli_compression"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Brotli Compression support",
		}
		blockAttributes["domain_inspector"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Domain Inspector support",
		}
		blockAttributes["image_optimizer"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Image Optimizer support (requires at least one backend with a `shield` attribute)",
		}
		blockAttributes["origin_inspector"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable Origin Inspector support",
		}
	}

	// websockets is supported for both Compute (wasm) and Deliver (vcl) services.
	blockAttributes["websockets"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
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
func (h *ProductEnablementServiceAttributeHandler) Create(_ context.Context, d *schema.ResourceData, resource map[string]any, _ int, conn *gofastly.Client) error {
	serviceID := d.Id()

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		if resource["fanout"].(bool) {
			log.Println("[DEBUG] fanout set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductFanout,
				ServiceID: serviceID,
			})
			if err != nil {
				return fmt.Errorf("failed to enable fanout: %w", err)
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
				return fmt.Errorf("failed to enable brotli_compression: %w", err)
			}
		}

		if resource["domain_inspector"].(bool) {
			log.Println("[DEBUG] domain_inspector set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductDomainInspector,
				ServiceID: serviceID,
			})
			if err != nil {
				return fmt.Errorf("failed to enable domain_inspector: %w", err)
			}
		}

		if resource["image_optimizer"].(bool) {
			log.Println("[DEBUG] image_optimizer set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductImageOptimizer,
				ServiceID: serviceID,
			})
			if err != nil {
				return fmt.Errorf("failed to enable image_optimizer: %w", err)
			}
		}

		if resource["origin_inspector"].(bool) {
			log.Println("[DEBUG] origin_inspector set")
			_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductOriginInspector,
				ServiceID: serviceID,
			})
			if err != nil {
				return fmt.Errorf("failed to enable origin_inspector: %w", err)
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
			return fmt.Errorf("failed to enable websockets: %w", err)
		}
	}

	return nil
}

// Read refreshes the resource.
func (h *ProductEnablementServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, _ int, conn *gofastly.Client) error {
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
			"name": "products",
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
func (h *ProductEnablementServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, _, modified map[string]any, _ int, conn *gofastly.Client) error {
	serviceID := d.Id()

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		if v, ok := modified["fanout"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] fanout set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductFanout,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to enable fanout: %w", err)
				}
			} else {
				log.Println("[DEBUG] fanout not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductFanout,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to disable fanout: %w", err)
				}
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		// FIXME: Looks like `modified` contains products that haven't been updated.
		// The only practical issue here is that an unnecessary API request is made.
		if v, ok := modified["brotli_compression"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] brotli_compression set")
				_, err := conn.EnableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductBrotliCompression,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to enable brotli_compression: %w", err)
				}
			} else {
				log.Println("[DEBUG] brotli_compression not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductBrotliCompression,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to disable brotli_compression: %w", err)
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
					return fmt.Errorf("failed to enable domain_inspector: %w", err)
				}
			} else {
				log.Println("[DEBUG] domain_inspector not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductDomainInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to disable domain_inspector: %w", err)
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
					return fmt.Errorf("failed to enable image_optimizer: %w", err)
				}
			} else {
				log.Println("[DEBUG] image_optimizer not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductImageOptimizer,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to disable image_optimizer: %w", err)
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
					return fmt.Errorf("failed to enable origin_inspector: %w", err)
				}
			} else {
				log.Println("[DEBUG] origin_inspector not set")
				err := conn.DisableProduct(&gofastly.ProductEnablementInput{
					ProductID: gofastly.ProductOriginInspector,
					ServiceID: serviceID,
				})
				if err != nil {
					return fmt.Errorf("failed to disable origin_inspector: %w", err)
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
				return fmt.Errorf("failed to enable websockets: %w", err)
			}
		} else {
			log.Println("[DEBUG] websockets not set")
			err := conn.DisableProduct(&gofastly.ProductEnablementInput{
				ProductID: gofastly.ProductWebSockets,
				ServiceID: serviceID,
			})
			if err != nil {
				return fmt.Errorf("failed to disable websockets: %w", err)
			}
		}
	}

	return nil
}

// Delete deletes the resource.
//
// IMPORTANT: We must allow a user to clean-up their state.
// If a user doesn't have self-enablement for a particular product and they add
// the product_enablement block to their config, then they'll have errors trying
// to enable that product. The problem now is if they remove the block from
// their config completely, they still won't be able to successfully apply the
// deletion because the API will error telling them they're not entitled to
// disable a product. So if that's the error we're getting back from the API,
// then we'll skip the error as we want the `terraform apply` to be successful
// and for the user to end up with a clean state.
//
// NOTE: We avoid returning early because there are multiple API calls.
// For example, if the first API call to disable a product failed because the
// user didn't have entitlement to disable, then returning either the error or
// skipping it and returning nil would cause the Delete function to finish and
// we wouldn't have a chance to disable the other products.
//
// TODO: Consider switching from a TypeSet to avoid unnecessary API calls.
// In a scenario where a new product is set to `true` (e.g. to be enabled) the
// set hash changes and so the set 'as a whole' is deleted (causing all the
// products to be disabled) and then all the APIs are called again to re-enable
// the products (even though they might not have actually been set to `false` in
// the first place). The solution would be to swap TypeSet for TypeList but then
// that means we'd lose the 'modified' data diff abstraction that was built into
// the Fastly provider. Look at the `package` block as an example of a TypeList,
// and you'll see it doesn't implement standard CRUD methods but has a single
// `Process` method that handles both CREATE and UPDATE stages and doesn't get
// passed a data structure that indicates what has changed like we do with the
// TypeSet data type. So it'll be a trade-off.
func (h *ProductEnablementServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, _ map[string]any, _ int, conn *gofastly.Client) error {
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		log.Println("[DEBUG] disable fanout")
		err := conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductFanout,
			ServiceID: d.Id(),
		})
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		log.Println("[DEBUG] disable brotli_compression")
		err := conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductBrotliCompression,
			ServiceID: d.Id(),
		})
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable domain_inspector")
		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductDomainInspector,
			ServiceID: d.Id(),
		})
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable image_optimizer")

		// IMPORTANT: There is no public API for checking user entitlement.
		// This means if a user adds product_enablement and realises they can't
		// enable products, then they'll need to delete the block from their config
		// and that will cause this Delete function to be run. The problem now, is
		// that the Product Enablement API allows you to disable the IO product even
		// if you don't have entitlement to 'enable' it (this is because we need to
		// allow a user to disable IO before their 'pro' free trial auto-renews).
		// The reason this is a problem is because a user might have requested
		// Fastly Support add IO to their service and now Terraform will go ahead
		// and disable it by accident. The only way to prevent that from happening
		// is to make an additional API call (for IO only) to try and 'enable' IO
		// and then if it fails with the error about missing entitlement we know we
		// can skip trying to 'disable' it.
		_, err = conn.EnableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductImageOptimizer,
			ServiceID: d.Id(),
		})
		if err != nil {
			if he, ok := err.(*gofastly.HTTPError); ok {
				if he.StatusCode == http.StatusBadRequest {
					for _, e := range he.Errors {
						if !strings.Contains(e.Title, "not entitled to enable") && !strings.Contains(e.Title, "product cannot be self enabled") {
							// NOTE: If this user is not entitled to disable IO it's still OK.
							// For example, if the user has added product_enablement by
							// accident (they don't have entitlement), then this call to
							// disable the product will not get executed and the rest of the
							// Delete method will run through until completion and because we
							// ignore errors returned from the API when deleting it means the
							// other calls to disable the other products won't cause the
							// Delete method to fail and thus allows the user to clean-up
							// their Terraform state.
							err = conn.DisableProduct(&gofastly.ProductEnablementInput{
								ProductID: gofastly.ProductImageOptimizer,
								ServiceID: d.Id(),
							})
							if err != nil {
								if e := h.checkAPIError(err); e != nil {
									return e
								}
							}
						}
					}
				}
			}
		}

		log.Println("[DEBUG] disable origin_inspector")
		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductOriginInspector,
			ServiceID: d.Id(),
		})
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}
	}

	log.Println("[DEBUG] disable websockets")
	err := conn.DisableProduct(&gofastly.ProductEnablementInput{
		ProductID: gofastly.ProductWebSockets,
		ServiceID: d.Id(),
	})
	if err != nil {
		if e := h.checkAPIError(err); e != nil {
			return e
		}
	}

	return nil
}

// checkAPIError inspects the error type for a title that has a message
// indicating the user cannot call the API because they are not entitled. For
// these users we want to skip the error so that we can allow them to clean up
// their Terraform state.
func (h *ProductEnablementServiceAttributeHandler) checkAPIError(err error) error {
	if he, ok := err.(*gofastly.HTTPError); ok {
		if he.StatusCode == http.StatusBadRequest {
			for _, e := range he.Errors {
				if strings.Contains(e.Title, "not entitled to disable") || strings.Contains(e.Title, "product cannot be disabled") {
					return nil
				}
			}
		}
	}
	return err
}
