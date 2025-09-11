resource "fastly_service_vcl" "example" {
  name = "demofastly"
  domain {
    name = "demo.notexample.com"
  }
}

resource "fastly_integration" "example" {
  name = "my_integration"
  type = "a_valid_type"

  config = {
    webhook = "some_webhook"
  }
}

resource "fastly_alert" "example" {
  name       = "my_vcl_service errors"
  service_id = fastly_service_vcl.example.id
  source     = "stats"
  metric     = "status_5xx"

  evaluation_strategy {
    type      = "above_threshold"
    period    = "5m"
    threshold = 10
  }

  integration_ids = [fastly_integration.example.id]
}
