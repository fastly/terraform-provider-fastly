package fastly

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/botmanagement"
	"github.com/fastly/go-fastly/v9/fastly/products/brotlicompression"
	"github.com/fastly/go-fastly/v9/fastly/products/ddosprotection"
	"github.com/fastly/go-fastly/v9/fastly/products/domaininspector"
	"github.com/fastly/go-fastly/v9/fastly/products/fanout"
	"github.com/fastly/go-fastly/v9/fastly/products/imageoptimizer"
	"github.com/fastly/go-fastly/v9/fastly/products/logexplorerinsights"
	"github.com/fastly/go-fastly/v9/fastly/products/ngwaf"
	"github.com/fastly/go-fastly/v9/fastly/products/origininspector"
	"github.com/fastly/go-fastly/v9/fastly/products/websockets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			Optional:    true,
			Default:     "products",
			Description: "Used by the provider to identify modified settings (changing this value will force the entire block to be deleted, then recreated)",
		},
	}

	// These products are supported only on Compute (WASM) services.
	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		blockAttributes["fanout"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Fanout support",
		}
	}

	// These products are supported only on Delivery (VCL) services.
	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		blockAttributes["bot_management"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Bot Management support",
		}
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
			Description: "Enable Image Optimizer support (all backends must have a `shield` attribute)",
		}
		blockAttributes["origin_inspector"] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Enable Origin Inspector support",
		}
	}

	// These products are supported for both Compute (WASM) and Delivery (VCL) services.
	blockAttributes["websockets"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Enable WebSockets support",
	}
	blockAttributes["log_explorer_insights"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Enable Log Explorer & Insights",
	}
	blockAttributes["ddos_protection"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "DDoS Protection product",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Enable DDoS Protection support",
				},
				"mode": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Operation mode",
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
						[]string{"off", "log", "block"},
						false,
					)),
				},
			},
		},
	}
	blockAttributes["ngwaf"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Next-Gen WAF product",
		MaxItems:    1,
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Enable Next-Gen WAF support",
				},
				"traffic_ramp": {
					Type:         schema.TypeInt,
					Optional:     true,
					Default:      100,
					Description:  "The percentage of traffic to inspect",
					ValidateFunc: validation.IntBetween(1, 100),
				},
				"workspace_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The workspace to link",
				},
			},
		},
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
			_, err := fanout.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable fanout: %w", err)
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if resource["bot_management"].(bool) {
			log.Println("[DEBUG] bot_management set")
			_, err := botmanagement.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable bot_management: %w", err)
			}
		}
		if resource["brotli_compression"].(bool) {
			log.Println("[DEBUG] brotli_compression set")
			_, err := brotlicompression.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable brotli_compression: %w", err)
			}
		}

		if resource["domain_inspector"].(bool) {
			log.Println("[DEBUG] domain_inspector set")
			_, err := domaininspector.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable domain_inspector: %w", err)
			}
		}

		if resource["image_optimizer"].(bool) {
			log.Println("[DEBUG] image_optimizer set")
			_, err := imageoptimizer.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable image_optimizer: %w", err)
			}
		}

		if resource["origin_inspector"].(bool) {
			log.Println("[DEBUG] origin_inspector set")
			_, err := origininspector.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable origin_inspector: %w", err)
			}
		}
	}

	if resource["websockets"].(bool) {
		log.Println("[DEBUG] websockets set")
		_, err := websockets.Enable(conn, serviceID)
		if err != nil {
			return fmt.Errorf("failed to enable websockets: %w", err)
		}
	}

	if resource["log_explorer_insights"].(bool) {
		log.Println("[DEBUG] log_explorer_insights set")
		_, err := logexplorerinsights.Enable(conn, serviceID)
		if err != nil {
			return fmt.Errorf("failed to enable log_explorer_insights: %w", err)
		}
	}

	ddp := resource["ddos_protection"].([]any)
	if len(ddp) != 0 {
		if ddp[0].(map[string]any)["enabled"].(bool) {
			log.Println("[DEBUG] ddos_protection set")
			_, err := ddosprotection.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable ddos_protection: %w", err)
			}

			// The operation mode is set by default to "log"
			mode := ddp[0].(map[string]any)["mode"].(string)
			if mode != "log" {
				_, err := ddosprotection.UpdateConfiguration(conn, serviceID, ddosprotection.ConfigureInput{
					Mode: mode,
				})
				if err != nil {
					return fmt.Errorf("failed to set the configuration of ddos_protection: %w", err)
				}
			}
		}
	}

	ngw := resource["ngwaf"].([]any)
	if len(ngw) != 0 {
		if ngw[0].(map[string]any)["enabled"].(bool) {
			log.Println("[DEBUG] ngwaf set")

			id := ngw[0].(map[string]any)["workspace_id"].(string)
			_, err := ngwaf.Enable(conn, serviceID, ngwaf.EnableInput{
				WorkspaceID: id,
			})
			if err != nil {
				return fmt.Errorf("failed to enable ngwaf: %w", err)
			}

			// The percentage of traffic to inspect is set by default to 100
			tr := ngw[0].(map[string]any)["traffic_ramp"].(int)
			if tr != 100 {
				_, err := ngwaf.UpdateConfiguration(conn, serviceID, ngwaf.ConfigureInput{
					WorkspaceID: id,
					TrafficRamp: strconv.Itoa(tr),
				})
				if err != nil {
					return fmt.Errorf("failed to set the configuration of ngwaf: %w", err)
				}
			}
		}
	}

	return nil
}

// Read refreshes the resource.
func (h *ProductEnablementServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, _ map[string]any, _ int, conn *gofastly.Client) error {
	localState := d.Get(h.Key()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		serviceID := d.Id()
		log.Printf("[DEBUG] Refreshing Product Enablement Configuration for (%s)", serviceID)

		// The API returns a 400 if a product is not enabled.
		// The API client returns an error if a non-2xx is returned from the API.

		result := map[string]any{}

		// The `name` attribute in this resource is used by default as a key for calculating diffs.
		// This is handled as part of the internal abstraction logic.
		//
		// See the call ToServiceAttributeDefinition() inside NewServiceProductEnablement()
		// See also the diffing logic:
		//   - https://github.com/fastly/terraform-provider-fastly/blob/4b9506fba1fd17e2bf760f447cbd8c394bb1e153/fastly/service_crud_attribute_definition.go#L94
		//   - https://github.com/fastly/terraform-provider-fastly/blob/4b9506fba1fd17e2bf760f447cbd8c394bb1e153/fastly/diff.go#L108-L117
		//
		// Because the name can be set by a user, we first check if the resource
		// exists in their state, and if so we'll use the value assigned there. If
		// they've not explicitly defined a name in their config, then the default
		// value will be returned.
		if len(localState) > 0 {
			name := localState[0].(map[string]any)["name"].(string)
			result["name"] = name
		}

		if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
			if _, err := fanout.Get(conn, serviceID); err == nil {
				result["fanout"] = true
			}
		}

		if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
			if _, err := botmanagement.Get(conn, serviceID); err == nil {
				result["bot_management"] = true
			}

			if _, err := brotlicompression.Get(conn, serviceID); err == nil {
				result["brotli_compression"] = true
			}

			if _, err := domaininspector.Get(conn, serviceID); err == nil {
				result["domain_inspector"] = true
			}

			if _, err := imageoptimizer.Get(conn, serviceID); err == nil {
				result["image_optimizer"] = true
			}

			if _, err := origininspector.Get(conn, serviceID); err == nil {
				result["origin_inspector"] = true
			}
		}

		if _, err := websockets.Get(conn, serviceID); err == nil {
			result["websockets"] = true
		}

		if _, err := logexplorerinsights.Get(conn, serviceID); err == nil {
			result["log_explorer_insights"] = true
		}

		if _, err := ddosprotection.Get(conn, serviceID); err == nil {
			c, err := ddosprotection.GetConfiguration(conn, serviceID)
			if err != nil {
				return fmt.Errorf("error looking up DDoS Protection product configuration for (%s): %s", serviceID, err)
			}

			ddp := []map[string]any{}
			ddp = append(ddp, map[string]any{
				"enabled": true,
				"mode":    *c.Configuration.Mode,
			})

			result["ddos_protection"] = ddp
		} else {
			if len(localState) > 0 {
				ddp := localState[0].(map[string]any)["ddos_protection"].([]any)
				result["ddos_protection"] = ddp
			}
		}

		if _, err := ngwaf.Get(conn, serviceID); err == nil {
			c, err := ngwaf.GetConfiguration(conn, serviceID)
			if err != nil {
				return fmt.Errorf("error looking up Next-Gen WAF product configuration for (%s): %s", serviceID, err)
			}

			tf, err := strconv.Atoi(*c.Configuration.TrafficRamp)
			if err != nil {
				return fmt.Errorf("error converting Next-Gen WAF's percentage of traffic for (%s): %s", serviceID, err)
			}

			ngw := []map[string]any{}
			ngw = append(ngw, map[string]any{
				"enabled":      true,
				"workspace_id": *c.Configuration.WorkspaceID,
				"traffic_ramp": tf,
			})

			result["ngwaf"] = ngw
		} else {
			if len(localState) > 0 {
				ngw := localState[0].(map[string]any)["ngwaf"].([]any)
				result["ngwaf"] = ngw
			}
		}

		results := []map[string]any{result}

		// IMPORTANT: Avoid runtime panic "set item just set doesn't exist".
		// TF will panic when trying to append an empty map to a TypeSet.
		// i.e. a typed nil.
		if len(results[0]) > 0 {
			if err := d.Set(h.Key(), results); err != nil {
				log.Printf("[WARN] Error setting Product Enablement for (%s): %s", serviceID, err)
				return err
			}
		}
	}

	return nil
}

// Update updates the resource.
//
// IMPORTANT: We ignore errors related to entitlement when updating.
//
// This is to provide a non-breaking workaround for customers who used an older
// version of the Fastly Terraform provider. See details in the PR:
// https://github.com/fastly/terraform-provider-fastly/pull/763
func (h *ProductEnablementServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, _, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	log.Println("[DEBUG] Update Product Enablement")

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		if v, ok := modified["fanout"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] fanout will be enabled")
				_, err := fanout.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable fanout: %w", err)
				}
			} else {
				log.Println("[DEBUG] fanout will be disabled")
				err := fanout.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if v, ok := modified["bot_management"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] bot_management will be enabled")
				_, err := botmanagement.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable bot_management: %w", err)
				}
			} else {
				log.Println("[DEBUG] bot_management will be disabled")
				err := botmanagement.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}

		if v, ok := modified["brotli_compression"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] brotli_compression will be enabled")
				_, err := brotlicompression.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable brotli_compression: %w", err)
				}
			} else {
				log.Println("[DEBUG] brotli_compression will be disabled")
				err := brotlicompression.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}

		if v, ok := modified["domain_inspector"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] domain_inspector will be enabled")
				_, err := domaininspector.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable domain_inspector: %w", err)
				}
			} else {
				log.Println("[DEBUG] domain_inspector will be disabled")
				err := domaininspector.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}

		if v, ok := modified["image_optimizer"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] image_optimizer will be enabled")
				_, err := imageoptimizer.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable image_optimizer: %w", err)
				}
			} else {
				log.Println("[DEBUG] image_optimizer will be disabled")
				err := imageoptimizer.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}

		if v, ok := modified["origin_inspector"]; ok {
			if v.(bool) {
				log.Println("[DEBUG] origin_inspector will be enabled")
				_, err := origininspector.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable origin_inspector: %w", err)
				}
			} else {
				log.Println("[DEBUG] origin_inspector will be disabled")
				err := origininspector.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}
	}

	if v, ok := modified["websockets"]; ok {
		if v.(bool) {
			log.Println("[DEBUG] websockets will be enabled")
			_, err := websockets.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable websockets: %w", err)
			}
		} else {
			log.Println("[DEBUG] websockets will be disabled")
			err := websockets.Disable(conn, serviceID)
			if err != nil {
				if e := h.checkAPIError(err); e != nil {
					return e
				}
			}
		}
	}

	if v, ok := modified["log_explorer_insights"]; ok {
		if v.(bool) {
			log.Println("[DEBUG] log_explorer_insights will be enabled")
			_, err := logexplorerinsights.Enable(conn, serviceID)
			if err != nil {
				return fmt.Errorf("failed to enable log_explorer_insights: %w", err)
			}
		} else {
			log.Println("[DEBUG] log_explorer_insights will be disabled")
			err := logexplorerinsights.Disable(conn, serviceID)
			if err != nil {
				if e := h.checkAPIError(err); e != nil {
					return e
				}
			}
		}
	}

	if v, ok := modified["ddos_protection"]; ok {
		ddp := v.([]any)
		if len(ddp) != 0 {
			if ddp[0].(map[string]any)["enabled"].(bool) {
				log.Println("[DEBUG] ddos_protection will be enabled")
				_, err := ddosprotection.Enable(conn, serviceID)
				if err != nil {
					return fmt.Errorf("failed to enable ddos_protection: %w", err)
				}

				// The operation mode is set by default to "log"
				mode := ddp[0].(map[string]any)["mode"].(string)
				if mode != "log" {
					log.Println("[DEBUG] ddos_protection mode will be updated")
					_, err := ddosprotection.UpdateConfiguration(conn, serviceID, ddosprotection.ConfigureInput{
						Mode: mode,
					})
					if err != nil {
						return fmt.Errorf("failed to set the configuration of ddos_protection: %w", err)
					}
				}
			} else {
				log.Println("[DEBUG] ddos_protection will be disabled")
				err := ddosprotection.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
			}
		}
	}

	if v, ok := modified["ngwaf"]; ok {
		ngw := v.([]any)
		if len(ngw) != 0 {
			if ngw[0].(map[string]any)["enabled"].(bool) {
				log.Println("[DEBUG] ngwaf will be enabled")

				id := ngw[0].(map[string]any)["workspace_id"].(string)
				_, err := ngwaf.Enable(conn, serviceID, ngwaf.EnableInput{
					WorkspaceID: id,
				})
				if err != nil {
					return fmt.Errorf("failed to enable ngwaf: %w", err)
				}

				tr := ngw[0].(map[string]any)["traffic_ramp"].(int)
				_, err = ngwaf.UpdateConfiguration(conn, serviceID, ngwaf.ConfigureInput{
					WorkspaceID: id,
					TrafficRamp: strconv.Itoa(tr),
				})
				if err != nil {
					return fmt.Errorf("failed to set the configuration of ngwaf: %w", err)
				}
			} else {
				log.Println("[DEBUG] ngwaf will be disabled")
				err := ngwaf.Disable(conn, serviceID)
				if err != nil {
					if e := h.checkAPIError(err); e != nil {
						return e
					}
				}
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
	serviceID := d.Id()

	if h.GetServiceMetadata().serviceType == ServiceTypeCompute {
		log.Println("[DEBUG] disable fanout")
		err := fanout.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}
	}

	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		log.Println("[DEBUG] disable bot_management")
		err := botmanagement.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable brotli_compression")
		err = brotlicompression.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable domain_inspector")
		err = domaininspector.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable image_optimizer")
		err = imageoptimizer.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}

		log.Println("[DEBUG] disable origin_inspector")
		err = origininspector.Disable(conn, serviceID)
		if err != nil {
			if e := h.checkAPIError(err); e != nil {
				return e
			}
		}
	}

	log.Println("[DEBUG] disable websockets")
	err := websockets.Disable(conn, serviceID)
	if err != nil {
		if e := h.checkAPIError(err); e != nil {
			return e
		}
	}

	log.Println("[DEBUG] disable log_explorer_insights")
	err = logexplorerinsights.Disable(conn, serviceID)
	if err != nil {
		if e := h.checkAPIError(err); e != nil {
			return e
		}
	}

	log.Println("[DEBUG] disable ddos_protection")
	err = ddosprotection.Disable(conn, serviceID)
	if err != nil {
		if e := h.checkAPIError(err); e != nil {
			return e
		}
	}

	log.Println("[DEBUG] disable ngwaf")
	err = ngwaf.Disable(conn, serviceID)
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
