package fastly

import (
	"context"
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ProductEnablementServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type ImageOptimizerDefaultSettingsServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

// NewServiceProductEnablement returns a new resource.
func NewServiceImageOptimizerDefaultSettings(sa ServiceMetadata) ServiceAttributeDefinition {
	return ToServiceAttributeDefinition(&ImageOptimizerDefaultSettingsServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "image_optimizer_default_settings",
			serviceMetadata: sa,
		},
	})
}

// Key returns the resource key.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Key() string {
	return h.key
}

// GetSchema returns the resource schema.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) GetSchema() *schema.Schema {
	attributes := map[string]*schema.Schema{
		"allow_video": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enables GIF to MP4 transformations on this service.",
		},
		"jpeg_quality": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      85,
			Description:  "The default quality to use with jpeg output. This can be overridden with the \"quality\" parameter on specific image optimizer requests.",
			ValidateFunc: validation.IntBetween(1, 100),
		},
		"jpeg_type": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "auto",
			Description:  "The default type of jpeg output to use. This can be overriden with \"format=bjpeg\" and \"format=pjpeg\" on specific image optimizer requests.",
			ValidateFunc: validation.StringInSlice([]string{"auto", "baseline", "progressive"}, false),
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "image_optimizer_default_settings",
			Description: "Used by the provider to identify modified settings (changing this value will force the entire block to be deleted, then recreated)",
		},
		"resize_filter": {
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "lanczos3",
			Description:  "The type of filter to use while resizing an image.",
			ValidateFunc: validation.StringInSlice([]string{"lanczos3", "lanczos2", "bicubic", "bilinear", "nearest"}, false),
		},
		"upscale": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether or not we should allow output images to render at sizes larger than input.",
		},
		"webp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Controls whether or not to default to webp output when the client supports it. This is equivalent to adding \"auto=webp\" to all image optimizer requests.",
		},
		"webp_quality": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      85,
			Description:  "The default quality to use with webp output. This can be overriden with the second option in the \"quality\" URL parameter on specific image optimizer requests",
			ValidateFunc: validation.IntBetween(1, 100),
		},
	}

	// NOTE: MaxItems: 1 (to enforce only one image_optimizer_default_settings per service).
	// lintignore:S018
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: attributes,
		},
	}
}

// Create creates the resource.
//
// If a user has Image Optimizer enabled, they will always have some default settings. So, creation and updating are synonymous.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Create(c context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	return h.Update(c, d, resource, resource, serviceVersion, conn)
}

// Read refreshes the resource.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Read(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	localState := d.Get(h.Key()).(*schema.Set).List()

	if len(localState) > 0 || d.Get("imported").(bool) || d.Get("force_refresh").(bool) {
		serviceID := d.Id()

		log.Printf("[DEBUG] Refreshing Image Optimizer default settings for (%s)", serviceID)

		remoteState, err := conn.GetImageOptimizerDefaultSettings(&gofastly.GetImageOptimizerDefaultSettingsInput{
			ServiceID:      serviceID,
			ServiceVersion: serviceVersion,
		})
		if err != nil {
			return err
		}
		// Handle the case where the service has no Image Optimizer default settings configured (for example, if it has never had Image
		// Optimizer enabled.)
		if remoteState == nil {
			return nil
		}

		result := map[string]any{
			"allow_video":   remoteState.AllowVideo,
			"jpeg_type":     remoteState.JpegType,
			"jpeg_quality":  remoteState.JpegQuality,
			"resize_filter": remoteState.ResizeFilter,
			"upscale":       remoteState.Upscale,
			"webp":          remoteState.Webp,
			"webp_quality":  remoteState.WebpQuality,
		}

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

		results := []map[string]any{result}

		if err := d.Set(h.Key(), results); err != nil {
			log.Printf("[WARN] Error setting Image Optimizer default setting for (%s): %s", d.Id(), err)
			return err
		}
	}

	return nil
}

// Update updates the resource.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Update(_ context.Context, d *schema.ResourceData, _, modified map[string]any, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	log.Println("[DEBUG] Update Image Optimizer default settings")

	if len(modified) == 0 {
		return nil
	}

	apiInput := gofastly.UpdateImageOptimizerDefaultSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}

	for key, value := range modified {
		switch key {
		case "resize_filter":
			var resizeFilter gofastly.ResizeFilter
			switch value.(string) {
			case "lanczos3":
				resizeFilter = gofastly.Lanczos3
			case "lanczos2":
				resizeFilter = gofastly.Lanczos2
			case "bicubic":
				resizeFilter = gofastly.Bicubic
			case "bilinear":
				resizeFilter = gofastly.Bilinear
			case "nearest":
				resizeFilter = gofastly.Nearest
			default:
				return fmt.Errorf("got unexpected resize_filter: %v", value)
			}
			apiInput.ResizeFilter = &resizeFilter
		case "webp":
			apiInput.Webp = gofastly.ToPointer(value.(bool))
		case "webp_quality":
			apiInput.WebpQuality = gofastly.ToPointer(value.(int))
		case "jpeg_type":
			apiInput.JpegType = gofastly.ToPointer(value.(string))
		case "jpeg_quality":
			apiInput.JpegQuality = gofastly.ToPointer(value.(int))
		case "upscale":
			apiInput.Upscale = gofastly.ToPointer(value.(bool))
		case "allow_video":
			apiInput.AllowVideo = gofastly.ToPointer(value.(bool))
		case "name":
			continue
		default:
			return fmt.Errorf("got unexpected image_optimizer_default_settings key: %v", key)
		}
	}

	log.Printf("[DEBUG] Calling Image Optimizer default settings update API with parameters: %+v", apiInput)

	if _, err := conn.UpdateImageOptimizerDefaultSettings(&apiInput); err != nil {
		return err
	}

	return nil
}

// Delete deletes the resource.
//
// Note: The API does not expose any way to reset default settings, so we don't have much to do here.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, _ map[string]any, _ int, conn *gofastly.Client) error {
	return nil
}
