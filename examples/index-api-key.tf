provider "fastly" {
  api_key = "test"
}

resource "fastly_service_v1" "myservice" {
  # ...
}