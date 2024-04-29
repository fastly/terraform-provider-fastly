package fastly

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

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
			Description:  "The default quality to use with JPEG output. This can be overridden with the \"quality\" parameter on specific image optimizer requests.",
			ValidateFunc: validation.IntBetween(1, 100),
		},
		"jpeg_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "auto",
			Description: "The default type of JPEG output to use. This can be overridden with \"format=bjpeg\" and \"format=pjpeg\" on specific image optimizer requests. Valid values are `auto`, `baseline` and `progressive`." + `
	- auto: Match the input JPEG type, or baseline if transforming from a non-JPEG input.
	- baseline: Output baseline JPEG images
	- progressive: Output progressive JPEG images`,
			ValidateFunc: validation.StringInSlice([]string{"auto", "baseline", "progressive"}, false),
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "image_optimizer_default_settings",
			Description: "Used by the provider to identify modified settings. Changing this value will force the entire block to be deleted, then recreated.",
		},
		"resize_filter": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "lanczos3",
			Description: "The type of filter to use while resizing an image. Valid values are `lanczos3`, `lanczos2`, `bicubic`, `bilinear` and `nearest`." + `
	- lanczos3: A Lanczos filter with a kernel size of 3. Lanczos filters can detect edges and linear features within an image, providing the best possible reconstruction.
	- lanczos2: A Lanczos filter with a kernel size of 2.
	- bicubic: A filter using an average of a 4x4 environment of pixels, weighing the innermost pixels higher.
	- bilinear: A filter using an average of a 2x2 environment of pixels.
	- nearest: A filter using the value of nearby translated pixel values. Preserves hard edges.`,
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
			Description: "Controls whether or not to default to WebP output when the client supports it. This is equivalent to adding \"auto=webp\" to all image optimizer requests.",
		},
		"webp_quality": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      85,
			Description:  "The default quality to use with WebP output. This can be overridden with the second option in the \"quality\" URL parameter on specific image optimizer requests.",
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
			var jpegType gofastly.JpegType
			switch value.(string) {
			case "auto":
				jpegType = gofastly.Auto
			case "baseline":
				jpegType = gofastly.Baseline
			case "progressive":
				jpegType = gofastly.Progressive
			default:
				return fmt.Errorf("got unexpected jpeg_type: %v", value)
			}
			apiInput.JpegType = &jpegType
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
// This resets Image Optimizer default settings to their defaults, to make it possible to easily undo any effect this block had.
//
// This assumes the service wasn't modified with the UI or any other non-terraform method. Given terraform's regular mode of operating within the world is to
// assume its in control of everything, I think that's quite a reasonable assumption.
func (h *ImageOptimizerDefaultSettingsServiceAttributeHandler) Delete(_ context.Context, d *schema.ResourceData, resource map[string]any, serviceVersion int, conn *gofastly.Client) error {
	serviceID := d.Id()
	log.Println("[DEBUG] Update Image Optimizer default settings")

	apiInput := gofastly.UpdateImageOptimizerDefaultSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}

	for key, value := range resource {
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
			var jpegType gofastly.JpegType
			switch value.(string) {
			case "auto":
				jpegType = gofastly.Auto
			case "baseline":
				jpegType = gofastly.Baseline
			case "progressive":
				jpegType = gofastly.Progressive
			default:
				return fmt.Errorf("got unexpected jpeg_type: %v", value)
			}
			apiInput.JpegType = &jpegType
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
		// inspect the error type for a title that has a message indicating the user cannot call the API because they do not have Image Optimizer
		// enabled. For these users we want to skip the error so that we can allow them to clean up their Terraform state. (also, because the Image Optimizer
		// default settings for services with Image Optimizer are effectively cleared by disabling Image Optimizer.)
		if he, ok := err.(*gofastly.HTTPError); ok {
			if he.StatusCode == http.StatusBadRequest {
				for _, e := range he.Errors {
					if strings.Contains(e.Detail, "Image Optimizer is not enabled on this service") {
						return nil
					}
				}
			}
		}

		return err

	}

	return nil
}
