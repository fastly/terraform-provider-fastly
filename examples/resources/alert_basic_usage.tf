resource "fastly_service_vcl" "example" {
  name = "my_vcl_service"
  # ...
}

resource "fastly_alert" "example" {
  name = "my_vcl_service errors"
  service_id = fastly_service_vcl.example.id
  source = "stats"
  metric = "status_5xx"

  evaluation_strategy {
    type = "above_threshold"
    period = "5m"
    threshold = 10
  }
}
