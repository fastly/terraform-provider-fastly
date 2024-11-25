---
layout: "fastly"
page_title: "Fastly: custom_dashboard"
sidebar_current: "docs-fastly-resource-custom_dashboard"
description: |-
  Provides a Custom Dashboard which can be viewed in the Fastly Web UI.
---

# fastly_custom_dashboard

Provides a Custom Dashboard which can be viewed in the Fastly Web UI.

## Example Usage

```terraform
resource "fastly_custom_dashboard" "example" {
  name = "My Cool Dashboard"
  description = "This dashboard has one chart"

  dashboard_item {
    title = "My Chart"
    subtitle = "This chart displays my cool metrics"

    span = 4

    data_source {
      type = "stats.edge"
      config {
        metrics = ["requests"]
      }
    }

    visualization {
      type = "chart"
      config {
        plot_type = "line"
        format = "number"
      }
    }
  }
}
```

## Import

Fastly Custom Dashboards can be imported using their ID, e.g.

```sh
$ terraform import fastly_custom_dashboard.example xxxxxxxxxxxxxxxxxxxx
```
