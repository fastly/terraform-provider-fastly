---
page_title: image_optimizer_default_settings
subcategory: "Guides"
---

## Image Optimizer Default Settings

The [Product Enablement](https://developer.fastly.com/reference/api/services/image-optimizer-default-settings/) API allows customers to configure default settings for Image Optimizer requests.

Not all customers are entitled to use these endpoints and so care needs to be given when configuring an `image_optimizer` block in your Terraform configuration.

Furthermore, even if a customer is entitled to use Image Optimizer, the product must also be enabled with `product_enablement`.

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

NOTE: Terraform will set all fields *regardless* of what's specified in the configuration. If a service was previously manually configured, adding this block will cause terraform to overwrite all unset settings with their default values.

## Delete

As Image Optimizer default settings will always have some value when a service has image optimizer enabled, deleting the `image_optimizer_default_settings`
block is a no-op.
