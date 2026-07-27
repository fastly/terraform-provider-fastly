package imageoptimizerdefaultsettings

import (
	"fmt"

	"github.com/fastly/terraform-provider-fastly/internal/service"

	fastly "github.com/fastly/go-fastly/v16/fastly"
)

func parseResizeFilter(v string) (*fastly.ImageOptimizerResizeFilter, error) {
	var rf fastly.ImageOptimizerResizeFilter
	switch v {
	case "lanczos3":
		rf = fastly.ImageOptimizerLanczos3
	case "lanczos2":
		rf = fastly.ImageOptimizerLanczos2
	case "bicubic":
		rf = fastly.ImageOptimizerBicubic
	case "bilinear":
		rf = fastly.ImageOptimizerBilinear
	case "nearest":
		rf = fastly.ImageOptimizerNearest
	default:
		return nil, fmt.Errorf("invalid resize_filter: %q", v)
	}
	return &rf, nil
}

func parseJpegType(v string) (*fastly.ImageOptimizerJpegType, error) {
	var jt fastly.ImageOptimizerJpegType
	switch v {
	case "auto":
		jt = fastly.ImageOptimizerAuto
	case "baseline":
		jt = fastly.ImageOptimizerBaseline
	case "progressive":
		jt = fastly.ImageOptimizerProgressive
	default:
		return nil, fmt.Errorf("invalid jpeg_type: %q", v)
	}
	return &jt, nil
}

// BuildUpdateInput builds the API input to fully replace Image Optimizer default settings
// for a service version. All fields are always populated so that boolean fields (which
// cannot be represented as "unset" in Go) are never accidentally dropped from an update.
func BuildUpdateInput(serviceID string, version int, m NestedModel) (*fastly.UpdateImageOptimizerDefaultSettingsInput, error) {
	resizeFilter, err := parseResizeFilter(service.StringValue(m.ResizeFilter))
	if err != nil {
		return nil, err
	}

	jpegType, err := parseJpegType(service.StringValue(m.JpegType))
	if err != nil {
		return nil, err
	}

	return &fastly.UpdateImageOptimizerDefaultSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
		AllowVideo:     new(service.BoolValue(m.AllowVideo)),
		JpegQuality:    new(int(service.Int64Value(m.JpegQuality))),
		JpegType:       jpegType,
		ResizeFilter:   resizeFilter,
		Upscale:        new(service.BoolValue(m.Upscale)),
		Webp:           new(service.BoolValue(m.Webp)),
		WebpQuality:    new(int(service.Int64Value(m.WebpQuality))),
	}, nil
}
