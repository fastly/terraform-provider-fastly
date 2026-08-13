  healthcheck {
    name = "{{.HEALTHCHECK_NAME}}"
    host = "example.com"
    path = "/healthz"
  }
