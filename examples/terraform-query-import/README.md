# terraform query + import Example

This example demonstrates how to use **`terraform query`** to discover existing
Fastly services and generate importable Terraform configuration with the
**dual-model Fastly Terraform provider**.

The generated `output.tf` was produced by running:

```bash
terraform query -generate-config-out=output.tf
```

It covers three services — `Prod-Service`, `Staging-Service`, and `Development-Service` — along with
their domains and backends as independent resources at the active version.

## What this example shows

- Using `terraform query` to list all `fastly_service_vcl_explicit`,
  `fastly_service_domain_explicit`, and `fastly_service_backend_explicit` resources
- Generated `import` blocks alongside each resource for use with `terraform apply`

## Files

```text
examples/terraform-query-import/
  main.tf                  # provider config
  fastly.tfquery.hcl       # list resource definitions for terraform query
  output.tf                # example output from terraform query -generate-config-out
```

## Workflow

```bash
export TF_CLI_CONFIG_FILE=.../bin/developer_overrides.tfrc
export FASTLY_API_KEY=<token>

# Discover all services and generate config
terraform query -generate-config-out=output.tf

# Review and apply (imports existing resources into state)
terraform plan
terraform apply
```
