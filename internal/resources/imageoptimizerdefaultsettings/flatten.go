package imageoptimizerdefaultsettings

import (
	"context"
	"net/http"
	"strings"

	fastly "github.com/fastly/go-fastly/v16/fastly"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func FlattenToNestedModel(s *fastly.ImageOptimizerDefaultSettings) NestedModel {
	if s == nil {
		return NestedModel{}
	}

	return NestedModel{
		AllowVideo:   types.BoolValue(s.AllowVideo),
		JpegQuality:  types.Int64Value(int64(s.JpegQuality)),
		JpegType:     types.StringValue(s.JpegType),
		ResizeFilter: types.StringValue(s.ResizeFilter),
		Upscale:      types.BoolValue(s.Upscale),
		Webp:         types.BoolValue(s.Webp),
		WebpQuality:  types.Int64Value(int64(s.WebpQuality)),
	}
}

// ReadForVersion reads Image Optimizer default settings for a service version.
//
// Because a service always has *some* default settings once Image Optimizer has ever been
// enabled (regardless of whether this block is configured in Terraform), the remote API is
// only queried when current is non-empty. This avoids surfacing settings as configuration
// drift for users who never declared this block.
func ReadForVersion(ctx context.Context, client *fastly.Client, serviceID string, version int, current []NestedModel) ([]NestedModel, error) {
	if len(current) == 0 {
		return nil, nil
	}

	remote, err := client.GetImageOptimizerDefaultSettings(ctx, &fastly.GetImageOptimizerDefaultSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return nil, err
	}
	if remote == nil {
		return nil, nil
	}

	return []NestedModel{FlattenToNestedModel(remote)}, nil
}

// Reconcile ensures the Image Optimizer default settings for a service version match desired.
//
// Image Optimizer default settings always exist server-side once Image Optimizer has been
// enabled, so create and update are the same operation: a full replace of all fields.
// Removing the block from configuration resets the settings back to their API defaults,
// but only when previous shows the block was actually configured before - otherwise there
// is nothing to reset, since the block was never under this resource's management.
func Reconcile(ctx context.Context, client *fastly.Client, serviceID string, version int, previous, desired []NestedModel) error {
	if len(desired) == 0 {
		if len(previous) == 0 {
			return nil
		}
		return resetToDefaults(ctx, client, serviceID, version)
	}

	remote, err := client.GetImageOptimizerDefaultSettings(ctx, &fastly.GetImageOptimizerDefaultSettingsInput{
		ServiceID:      serviceID,
		ServiceVersion: version,
	})
	if err != nil {
		return err
	}
	if remote != nil && desired[0].ModelsEqual(FlattenToNestedModel(remote)) {
		return nil
	}

	return update(ctx, client, serviceID, version, desired[0])
}

func update(ctx context.Context, client *fastly.Client, serviceID string, version int, m NestedModel) error {
	input, err := BuildUpdateInput(serviceID, version, m)
	if err != nil {
		return err
	}

	_, err = client.UpdateImageOptimizerDefaultSettings(ctx, input)
	return err
}

// resetToDefaults resets Image Optimizer default settings back to their API defaults. If the
// service no longer has Image Optimizer enabled, the API rejects the update; that error is
// swallowed since a service without Image Optimizer enabled already has no default settings.
func resetToDefaults(ctx context.Context, client *fastly.Client, serviceID string, version int) error {
	err := update(ctx, client, serviceID, version, defaultNestedModel())
	if err == nil {
		return nil
	}

	if he, ok := err.(*fastly.HTTPError); ok && he.StatusCode == http.StatusBadRequest {
		for _, e := range he.Errors {
			if strings.Contains(e.Detail, "Image Optimizer is not enabled on this service") {
				return nil
			}
		}
	}

	return err
}
