---
layout: "fastly"
page_title: "Fastly: user_v1"
sidebar_current: "docs-fastly-resource-user-v1"
description: |-
  Provides a Fastly User
---

# fastly_user_v1

Provides a Fastly User, representing the configuration for a user account for interacting with Fastly.

The User resource requires a login and name, and optionally a role.

## Example Usage

Basic usage:

```terraform
resource "fastly_user_v1" "demo" {
  login = "demo@example.com"
  name  = "Demo User"
}
```

## Import

A Fastly User can be imported using their user ID, e.g.

```sh
$ terraform import fastly_user_v1.demo xxxxxxxxxxxxxxxxxxxx
```