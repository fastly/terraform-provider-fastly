---
page_title: image_optimizer_default_settings
subcategory: "Guides"
---

## Image Optimizer Default Settings

[Fastly Image Optimizer](https://docs.fastly.com/products/image-optimizer) (Fastly IO) is an [image optimization](https://www.fastly.com/learning/what-is-image-optimization) service that manipulates and transforms your images in real time and caches optimized versions of them.

Fastly Image Optimizer supports a variety of image formats and applies specific settings to all images by default. These can be controlled with this API or the [web interface](https://docs.fastly.com/en/guides/about-fastly-image-optimizer#configuring-default-image-settings). Changes to other image settings, including most image transformations, require using query string parameters on individual requests.

The [Image Optimizer default settings](https://developer.fastly.com/reference/api/services/image-optimizer-default-settings/) API allows customers to configure
default settings for Image Optimizer requests, configuring the way images are optimized when not overridden by URL parameters on specific requests.

The service must have the Image Optimizer product enabled using the Product Enablement API, UI, or Terraform block to use the `image_optimizer` block.

## Example Usage

Basic usage:

```terraform
resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  product_enablement {
    image_optimizer = true
  }

  image_optimizer_default_settings {
    resize_filter = "lanczos3"
    webp = false
    webp_quality = 85
    jpeg_type = "auto"
    jpeg_quality = 85
    upscale = false
    allow_video = false
  }

  force_destroy = true
}
```

All fields in `image_optimizer_default_settings` are optional.

NOTE: When added, `image_optimizer_default_settings` will always set all default settings. This will replace any settings previously changed in the UI or API.

## Delete

Deleting the resource will reset all Image Optimizer default settings to their default values.

If deleting the resource errors due to Image Optimizer no longer being enabled on the service, then this error will be ignored.

When Image Optimizer is next re-enabled on a service, that service's Image Optimizer default settings will be reset - so a disabled service effectively already
has deleted/default Image Optimizer default settings.
