  healthcheck {
    name = "{{.HEALTHCHECK_NAME_1}}"
    host = "example.com"
    path = "/healthz"
  }

  healthcheck {
    name              = "{{.HEALTHCHECK_NAME_2}}"
    host              = "other.example.com"
    path              = "/status"
    check_interval    = 10000
    expected_response = 204
    headers           = ["X-Api-Key: abc123"]
    http_version      = "1.0"
    initial           = 1
    method            = "GET"
    threshold         = 2
    timeout           = 3000
    window            = 10
  }
