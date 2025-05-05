---
page_title: sensitive_attributes
subcategory: "Guides"
---

## Sensitive Attributes

When Terraform detects an attribute marked as Sensitive within a block, it will prevent that attribute from being displayed in the output. Instead, Terraform will comment:
“At least one attribute in this block is (or was) sensitive, so its contents will not be displayed.” 
This behavior can prevent users from viewing other attributes in the block that they may need, as well as hiding the values of the sensitive fields themselves.

## Displaying Sensitive Fields

The Fastly Terraform Provider allows you to override this behavior by setting the environment variable `FASTLY_TF_DISPLAY_SENSITIVE_FIELDS` to `true`. This will enable the display of sensitive field values within blocks that contain such attributes. To ensure this setting is applied, you must set the variable in an environment where your Go-based environment can access it—typically in the .zshrc or .bashrc file on your local machine, or as an environment variable in Docker containers.

## Warnings
Be cautious when enabling this setting in continuous integration (CI) environments. Exposing sensitive data in CI job logs can result in your data being inadvertently persisted in job records, which may lead to security risks.
