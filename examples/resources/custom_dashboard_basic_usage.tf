resource "fastly_custom_dashboard" "example" {
  name        = "Example Custom Dashboard"
  description = "This is an example custom dashboard. A few dashboard items are been provided to help you get started."

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
    subtitle = "Total errors served from the client of server."

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
