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
  name        = "Example Custom Dashboard"
  description = "This is an example custom dashboard. A few dashboard items are provided to help you get started."

  dashboard_item {
    title    = "Total Requests"
    subtitle = "Number of requests processed."

    data_source {
      type = "stats.edge"
      config = {
        metrics = ["requests"]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "requests"
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "Hit Ratio"
    subtitle = "Ratio of requests served from Fastly."

    data_source {
      type = "stats.edge"
      config {
        metrics = ["hit_ratio"]
      }
    }

    visualization {
      type = "chart"
      config {
        format             = "percent"
        plot_type          = "donut"
        calculation_method = "latest"
      }
    }
  }

  dashboard_item {
    title    = "Client & Server Errors"
    subtitle = "Total errors served from the client or server."

    data_source {
      type = "stats.edge"
      config {
        metrics = [
          "status_4xx",
          "status_5xx"
        ]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "requests"
        plot_type = "bar"
      }
    }
  }

  dashboard_item {
    title    = "Domains Requests"
    subtitle = "Requests by Domain."
    span     = 6

    data_source {
      type = "stats.domain"
      config {
        metrics = ["requests"]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "requests"
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "Origin Responses"
    subtitle = "Responses by Origin."
    span     = 6

    data_source {
      type = "stats.origin"
      config {
        metrics = ["all_responses"]
      }
    }

    visualization {
      type = "chart"
      config {
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "Total Bandwidth"
    subtitle = "Total bandwidth served."
    span     = 12

    data_source {
      type = "stats.edge"
      config {
        metrics = ["bandwidth"]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "bytes"
        plot_type = "bar"
      }
    }
  }

  dashboard_item {
    title    = "Products - Image Optimizer & Real-Time Log Streaming"
    subtitle = "Total IO images served and log statements sent."
    span     = 8

    data_source {
      type = "stats.edge"
      config {
        metrics = [
          "imgopto",
          "log"
        ]
      }
    }

    visualization {
      type = "chart"
      config {
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "Transport Protocols & Security"
    subtitle = "HTTP Protocols & TLS."

    data_source {
      type = "stats.edge"
      config {
        metrics = [
          "http1",
          "http2",
          "http3",
          "tls_v10",
          "tls_v11",
          "tls_v12",
          "tls_v13"
        ]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "requests"
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "Origin Miss Latency"
    subtitle = "Miss latency times for your origins."
    span     = 12

    data_source {
      type = "stats.edge"
      config {
        metrics = ["origin_latency"]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "milliseconds"
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "DDOS - Request Flood Attempts"
    subtitle = "Number of connections the limit-streams action was applied."
    span     = 6

    data_source {
      type = "stats.edge"
      config {
        metrics = [
          "ddos_action_limit_streams_connections",
          "ddos_action_limit_streams_requests"
        ]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "requests"
        plot_type = "line"
      }
    }
  }

  dashboard_item {
    title    = "DDOS - Malicious Bot Attack"
    subtitle = "Number of times the blackhole action was taken."
    span     = 6

    data_source {
      type = "stats.edge"
      config {
        metrics = [
          "ddos_action_close",
          "ddos_action_blackhole"
        ]
      }
    }

    visualization {
      type = "chart"
      config {
        format    = "number"
        plot_type = "line"
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

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A human-readable name.

### Optional

- `dashboard_item` (Block List, Max: 100) A list of dashboard items. (see [below for nested schema](#nestedblock--dashboard_item))
- `description` (String) A short description of the dashboard.

### Read-Only

- `id` (String) Dashboard identifier (UUID).

<a id="nestedblock--dashboard_item"></a>
### Nested Schema for `dashboard_item`

Required:

- `data_source` (Block List, Min: 1, Max: 1) An object which describes the data to display. (see [below for nested schema](#nestedblock--dashboard_item--data_source))
- `subtitle` (String) A human-readable subtitle for the dashboard item. Often a description of the visualization.
- `title` (String) A human-readable title for the dashboard item.
- `visualization` (Block List, Min: 1, Max: 1) An object which describes the data visualization to display. (see [below for nested schema](#nestedblock--dashboard_item--visualization))

Optional:

- `span` (Number) The number of columns for the dashboard item to span. Dashboards are rendered on a 12-column grid on "desktop" screen sizes.

<a id="nestedblock--dashboard_item--data_source"></a>
### Nested Schema for `dashboard_item.data_source`

Required:

- `config` (Block List, Min: 1, Max: 1) Configuration options for the selected data source. (see [below for nested schema](#nestedblock--dashboard_item--data_source--config))
- `type` (String) The source of the data to display. One of: `stats.edge`, `stats.domain`, `stats.origin`.

<a id="nestedblock--dashboard_item--data_source--config"></a>
### Nested Schema for `dashboard_item.data_source.config`

Required:

- `metrics` (List of String) The metrics to visualize. Valid options are defined by the selected data source: [stats.edge](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/edge/), [stats.domain](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/domain/), [stats.origin](https://www.fastly.com/documentation/reference/api/observability/custom-dashboards/metrics/origin/).



<a id="nestedblock--dashboard_item--visualization"></a>
### Nested Schema for `dashboard_item.visualization`

Required:

- `config` (Block List, Min: 1, Max: 1) Configuration options for the selected data source. (see [below for nested schema](#nestedblock--dashboard_item--visualization--config))
- `type` (String) The type of visualization to display. One of: `chart`.

<a id="nestedblock--dashboard_item--visualization--config"></a>
### Nested Schema for `dashboard_item.visualization.config`

Required:

- `plot_type` (String) The type of chart to display. One of: `line`, `bar`, `single-metric`, `donut`.

Optional:

- `calculation_method` (String) The aggregation function to apply to the dataset. One of: `avg`, `sum`, `min`, `max`, `latest`, `p95`.
- `format` (String) The units to use to format the data. One of: `number`, `bytes`, `percent`, `requests`, `responses`, `seconds`, `milliseconds`, `ratio`, `bitrate`.
