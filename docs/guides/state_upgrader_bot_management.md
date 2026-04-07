# State Upgrader for bot_management Schema Change

## Overview

This document describes the state upgrader implemented for the breaking change introduced in v9.0.0, where the `bot_management` attribute in the `product_enablement` block was changed from a boolean to a list block with nested attributes.

## The Change state_upgrader_bot_management

### Old Schema
```hcl
resource "fastly_service_vcl" "example" {
  # ...
  
  product_enablement {
    bot_management = true  # or false
  }
}
```

### New Schema
```hcl
resource "fastly_service_vcl" "example" {
  # ...
  
  product_enablement {
    bot_management {
      enabled      = true
      contentguard = "off"  # or "on"
    }
  }
}
```

## Implementation

The state upgrader automatically handles the migration of existing Terraform state files from the old schema to the new schema. This allows users to upgrade to version 9.0.0+ without manually editing their state files.

## Migration Logic

The upgrader performs the following conversions:

| Old State Value | New State Value |
|----------------|-----------------|
| `bot_management = true` | `bot_management { enabled = true, contentguard = "off" }` |
| `bot_management = false` | `bot_management = []` (empty list) |
| Already in new format | No change (idempotent) |

### Default Values

When migrating from `bot_management = true`, the upgrader sets:
- `enabled = true` (preserves the intent)
- `contentguard = "off"` (safe default, can be changed by user)

## Usage

The state upgrader runs automatically when users:

1. Upgrade to Terraform Provider Fastly v9.0.0+
2. Run `terraform plan` or `terraform apply`
3. Have existing state with the old `bot_management` boolean format

No manual intervention is required. Terraform will automatically detect that the state schema version is 0 and apply the upgrader to bring it to version 1.
