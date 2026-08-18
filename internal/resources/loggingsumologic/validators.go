package loggingsumologic

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// isURL rejects a string that is not an absolute http(s) URL with a host. The
// Fastly API accepts (and stores) whatever string is sent, so a malformed
// collector URL would otherwise only surface once logs silently fail to
// arrive at Sumo Logic, rather than at plan/validate time.
type isURL struct{}

func (isURL) Description(_ context.Context) string {
	return "value must be a valid absolute http:// or https:// URL"
}

func (v isURL) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (isURL) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	v := req.ConfigValue.ValueString()
	u, err := url.Parse(v)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid URL",
			"must be a valid absolute URL with an http or https scheme, e.g. \"https://collectors.sumologic.com/receiver/v1/http/<token>\".",
		)
	}
}
