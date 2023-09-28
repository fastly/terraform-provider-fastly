package fastly

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	gofastly "github.com/fastly/go-fastly/v8/fastly"
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

	// IMPORTANT: Avoid a diff unless state actually contains product_enablement.
	//
	// For all other nested blocks we go to the API to get the latest data if the
	// user has already configured the block in their config (hence it exists in
	// their state file) OR if the user is doing a `terraform import` OR if the
	// user has changed the active service version outside of Terraform and the
	// provider now needs to re-sync so it's tracking the correct version.
	//
	// But that approach means that users who don't configure a product_enablement
	// block but fall into the other two categories (i.e. importing or service
	// version drift) will unexpectedly see a `product_enablement` block in their
	// next plan/diff output and it will be suggesting that the block be deleted.
	//
	// This causes confusion for users and they're not sure how to resolve the
	// diff and sometimes will ADD the block into their config and try to set
	// different values, which then causes other errors as they may not be
	// entitled to actually use the Product Enablement API.
	//
	// To avoid this confusion, for product_enablement only, we only go to the API
	// if the user has either configured the product_enablement block in their
	// config or if they've configured it AND they're either importing or if there
	// is a force refresh due to the service version drifting.
	//
	// The reason we have to treat the Product Enablement (PE) API differently is
	// because its access is restricted to entitled users and so unlike other API
	// endpoints (e.g. backends or domains etc), including calls to the PE API can
	// cause unexpected diffs and runtime errors that stop a plan from being
	// applied successfully.
	existsInState := len(localState) > 0
	if existsInState || (existsInState && (d.Get("imported").(bool) || d.Get("force_refresh").(bool))) {
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
//
// IMPORTANT: The Update method is never called due to an implementation bug.
//
// It's a side-effect of the SetDiff logic https://github.com/fastly/terraform-provider-fastly/blob/6e03cce3127a30db94c1cdfa8eb621d9a7231989/fastly/service_crud_attribute_definition.go#L89-L95
// The SetDiff logic uses `name` as a computed key (i.e. if there’s a change to
// `name`, then it means the resource has changed and needs to be recreated).
//
// The problem is we defined `name` as a Computed attribute, which means it will
// always be marked as being changed as it's reset to an empty string due to it
// being a Computed attribute where the value is known only AFTER an apply.
//
// Fixing this bug would mean needing to change `name` from being Computed to
// Optional and also setting a default to match the hardcoded value "products"
// that we used when it was Computed, and to configure the attribute to be
// ignored by the Terraform diff processing logic. That should in theory make it
// a non-breaking change.
//
// This isn't the end of the world, it just means there are a few more API calls
// made. We're also (as of Sept 2023) in the process of rewriting the Terraform
// provider and so it might be best to resolve this as part of the rewrite.
func (h *ProductEnablementServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, _, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
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
// we wouldn't have a chance to attempt disabling the other products which they
// might be allowed to disable.
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
		err = conn.DisableProduct(&gofastly.ProductEnablementInput{
			ProductID: gofastly.ProductImageOptimizer,
			ServiceID: d.Id(),
		})
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
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
				if strings.Contains(e.Title, "not entitled to disable") || strings.Contains(e.Title, "product cannot be disabled") || strings.Contains(e.Title, "cannot self-disable") {
					return nil
				}
			}
		}
	}
	return err
}
