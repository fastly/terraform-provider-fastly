package imageoptimizerdefaultsettings

import (
	"github.com/fastly/terraform-provider-fastly/internal/service"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DefaultAllowVideo   = false
	DefaultJpegQuality  = 85
	DefaultJpegType     = "auto"
	DefaultResizeFilter = "lanczos3"
	DefaultUpscale      = false
	DefaultWebp         = false
	DefaultWebpQuality  = 85
)

var (
	JpegTypes     = []string{"auto", "baseline", "progressive"}
	ResizeFilters = []string{"lanczos3", "lanczos2", "bicubic", "bilinear", "nearest"}
)

type NestedModel struct {
	AllowVideo   types.Bool   `tfsdk:"allow_video"`
	JpegQuality  types.Int64  `tfsdk:"jpeg_quality"`
	JpegType     types.String `tfsdk:"jpeg_type"`
	ResizeFilter types.String `tfsdk:"resize_filter"`
	Upscale      types.Bool   `tfsdk:"upscale"`
	Webp         types.Bool   `tfsdk:"webp"`
	WebpQuality  types.Int64  `tfsdk:"webp_quality"`
}

func (n NestedModel) ModelsEqual(other NestedModel) bool {
	return service.BoolValue(n.AllowVideo) == service.BoolValue(other.AllowVideo) &&
		service.Int64Value(n.JpegQuality) == service.Int64Value(other.JpegQuality) &&
		service.StringValue(n.JpegType) == service.StringValue(other.JpegType) &&
		service.StringValue(n.ResizeFilter) == service.StringValue(other.ResizeFilter) &&
		service.BoolValue(n.Upscale) == service.BoolValue(other.Upscale) &&
		service.BoolValue(n.Webp) == service.BoolValue(other.Webp) &&
		service.Int64Value(n.WebpQuality) == service.Int64Value(other.WebpQuality)
}

func defaultNestedModel() NestedModel {
	return NestedModel{
		AllowVideo:   types.BoolValue(DefaultAllowVideo),
		JpegQuality:  types.Int64Value(DefaultJpegQuality),
		JpegType:     types.StringValue(DefaultJpegType),
		ResizeFilter: types.StringValue(DefaultResizeFilter),
		Upscale:      types.BoolValue(DefaultUpscale),
		Webp:         types.BoolValue(DefaultWebp),
		WebpQuality:  types.Int64Value(DefaultWebpQuality),
	}
}

func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"allow_video": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultAllowVideo),
			Description: "Enables GIF to MP4 transformations on this service. Default `false`.",
		},
		"jpeg_quality": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultJpegQuality),
			Description: "The default quality to use with JPEG output. This can be overridden with the `quality` parameter on specific image optimizer requests. Default `85`.",
			Validators: []validator.Int64{
				int64validator.Between(1, 100),
			},
		},
		"jpeg_type": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultJpegType),
			Description: "The default type of JPEG output to use. This can be overridden with `format=bjpeg` and `format=pjpeg` on specific image optimizer requests. Valid values are `auto`, `baseline` and `progressive`. Default `auto`.",
			Validators: []validator.String{
				stringvalidator.OneOf(JpegTypes...),
			},
		},
		"resize_filter": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(DefaultResizeFilter),
			Description: "The type of filter to use while resizing an image. Valid values are `lanczos3`, `lanczos2`, `bicubic`, `bilinear` and `nearest`. Default `lanczos3`.",
			Validators: []validator.String{
				stringvalidator.OneOf(ResizeFilters...),
			},
		},
		"upscale": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultUpscale),
			Description: "Whether or not we should allow output images to render at sizes larger than input. Default `false`.",
		},
		"webp": schema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(DefaultWebp),
			Description: "Controls whether or not to default to WebP output when the client supports it. This is equivalent to adding `auto=webp` to all image optimizer requests. Default `false`.",
		},
		"webp_quality": schema.Int64Attribute{
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(DefaultWebpQuality),
			Description: "The default quality to use with WebP output. This can be overridden with the second option in the `quality` URL parameter on specific image optimizer requests. Default `85`.",
			Validators: []validator.Int64{
				int64validator.Between(1, 100),
			},
		},
	}
}

// NestedBlockSchema returns the image_optimizer_default_settings block for use inside
// _auto aggregate resources. At most one block is supported per service, since Image
// Optimizer default settings are a singleton per service version.
func NestedBlockSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Image Optimizer default settings for this service. At most one block is supported.",
		NestedObject: schema.NestedBlockObject{
			Attributes: CommonAttributes(),
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func Equal(a, b []NestedModel) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	return a[0].ModelsEqual(b[0])
}
