---
page_title: image_optimizer_default_settings
subcategory: "Guides"
---

## Image Optimizer Default Settings

[Fastly Image Optimizer](https://docs.fastly.com/products/image-optimizer) (Fastly IO) is an [image optimization](https://www.fastly.com/learning/what-is-image-optimization) service that manipulates and transforms your images in real time and caches optimized versions of them.

Fastly Image Optimizer supports a variety of image formats and applies specific settings to all images by default. These can be controlled with this API or the [web interface](https://docs.fastly.com/en/guides/about-fastly-image-optimizer#configuring-default-image-settings). Changes to other image settings, including most image transformations, require using query string parameters on individual requests.

The [Image Optimizer default settings](https://developer.fastly.com/reference/api/services/image-optimizer-default-settings/) API allows customers to configure
default settings for Image Optimizer requests, configuring the way images are optimized when not overridden by URL parameters on specific requests.

Not all customers are entitled to use these endpoints and so care needs to be given when configuring an `image_optimizer` block in your Terraform configuration.

Image Optimizer must also be enabled on the specific service. This can be achieved in terraform with the `product_enablement` block.

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

As Image Optimizer default settings will always have some value when a service has image optimizer enabled, deleting the `image_optimizer_default_settings`
block is a no-op.
