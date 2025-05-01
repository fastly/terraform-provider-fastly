---
page_title: sensitive_attributes
subcategory: "Guides"
---

## Sensitive Attributes

If Terraform detects a attribute with the `Sensitive` attribute inside of a block, it will not display any attribute from that block in the output, instead commenting `At least one attribute in this block is (or was) sensitive, so its contents will not be displayed.` This can prevent usrs from being able to see other attributes that they want to, or even the values of the sensitive fields themselves.

## Displaying Sensitive Fields

The Fastly Terraform Provider can suppress this warning and display the information in all blocks containing attributes marked `Sensitive` by setting the environment variable `FASTLY_TF_DISPLAY_SENSITIVE_FIELDS` to true in a place where your Golang env can read it (usually the `.zshrc` or `.bashrc` on a local machine or in env vars for Docker containers).

## Warnings
It is not advised to set this variable on CI jobs as it could lead to your data being persisted on those CI job records
