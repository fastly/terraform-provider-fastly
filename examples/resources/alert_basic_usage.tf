resource "fastly_service_vcl" "myservice" {
  #...
}

resource "fastly_alert" "demo" {
  name = "myservice error rate"
  service_id = fastly_service_vcl.myservice.id
  source = "stats"
  metric = "status_5xx_rate"
  evaluation_strategy {
    type = "above_threshold"
    period = "5m"
    threshold = 0.1
  }
}
