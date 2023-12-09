---
page_title: product_enablement
subcategory: "Guides"
---

## Product Enablement

The [Product Enablement](https://developer.fastly.com/reference/api/products/enablement/) APIs allow customers to enable and disable specific products.

Not all customers are entitled to use these endpoints and so care needs to be given when configuring a `product_enablement` block in your Terraform configuration.

## Create

When defining the `product_enablement` block in either your `fastly_service_compute` or `fastly_service_vcl` resource, if you set one attribute (e.g. `brotli_compression`), then the Create function inside the provider will check if the attribute value is set to `true`.

Only if the product attribute is set to `true` will the provider attempt to enable the product.

If a product is not defined in your Terraform config, it is assigned a default value of `false`, meaning none of those products will be 'disabled'.

## Read

The Read function is used to update the Terraform state file. The Terraform provider calls the Fastly API to identify if a product is enabled or not, and it will assign the relevant `true` or `false` values to the state file.

If the Terraform configuration doesn't align with what the state file determines is how things are set up in the 'real world', then the Terraform plan/apply will indicate this in its diff when it compares the state file to your config file.

## Update

The Update function checks to see if the product has a different value to what it was set to originally (e.g. have you changed a product from `true` to `false` or vice-versa) and if it has been modified, then the value assigned to a product is compared to see if it's set to `true` or `false`.

If it's set to `true`, then the provider will attempt to call the Fastly API to 'enable' that product. Otherwise, if the value is set to `false`, then the provider will attempt to call the Fastly API to 'disable' that product.

The important part to pay attention to here is the 'entitlement' to call the Fastly API. Some customers don't have access to programmatically enable/disable products. Products have to then be set up manually by Fastly customer support.

If you _do_ have programmatic access to the [Product Enablement](https://developer.fastly.com/reference/api/products/enablement/) APIs, then you should ensure the correct value is assigned in your Terraform configuration to avoid accidentally disabling a product.

If you're unsure about whether you have API access, then we recommend reaching out to [support@fastly.com](mailto:support@fastly.com) to have them review your account settings.

## Delete

When deleting the `product_enablement` block from your Terraform configuration, the Delete function inside the provider will attempt to disable each product.

If the API returns an error, then that error will be returned to the user and consequently the `terraform apply` will fail. But, the error is only returned if it is _not_ an error related to permissions/access to the API endpoint itself.

This means, that if you delete the `product_enablement` block and you don't have access to disable Image Optimizer, then the error that is returned from the API to the Terraform provider will indicate that it failed because you don't have access to disable the product and that particular error will be ignored by the provider, subsequently allowing the `terraform apply` to complete successfully.
