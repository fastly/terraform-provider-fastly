<div align="center">
  <h3 align="center">Fastly Terraform Provider Issues</h3>
  <p align="center">Best practices for submitting an issue to the Fastly Terraform Provider repository.</p>
</div>

## Issue Type: Bug

Issues related to the Fastly Terraform Provider behavior not working as intended. 

- The Terraform Provider crashes or exits with an unexpected error
- An item in the Terraform Provider documentation is incorrect
- I am experiencing a drift after a `terraform apply` when I made no configuration changes

**Example:** "The provider crashes after I run a `terraform import`."

## Issue Type: Feature Request

Issues related to suggesting improvements to the Fastly Terraform Provider:

- Add support for an existing Fastly API endpoint / feature that doesn't exist in the provider
- Improved error messages or user experience
- Improved documentation around a specific feature

**Example:** "Add support for a LogABC resource and data source, which already exists in the Fastly API."

## Fastly Support

Fastly Terraform Provider behavior specific to your environment or service / account should be routed to the Fastly support team @ support.fastly.com or support@fastly.com. 

- A feature is missing from your account / service
- Partial content is returned that you may not have access to with your current Fastly account role
- My site is not loading after a configuration change

**Example:** After running `terraform apply`, an error is thrown that a given version can't be activated.
