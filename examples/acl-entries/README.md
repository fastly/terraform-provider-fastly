# ACL Entries Example

This example demonstrates managing a Fastly ACL and its entries using the `fastly_acl` and `fastly_acl_entries` resources.

## Usage

1. Set your Fastly API token:
   ```bash
   export FASTLY_API_TOKEN=your_token_here
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

## Features

- Creates an ACL, independent of any service or service version
- Manages CIDR-based allow/block entries in the ACL

## Important Notes

- `fastly_acl` is versionless: it is not tied to a service version
- The `manage_entries` attribute controls whether Terraform should detect and reconcile drift in ACL entries
- Set `manage_entries = false` (the default) if entries are managed externally (e.g., via API or other tools)
- Entries are keyed by CIDR prefix; each value must be `ALLOW` or `BLOCK`
