  header {
    name        = "{{.HEADER_NAME}}"
    action      = "delete"
    type        = "cache"
    destination = "http.x-amz-request-id"
  }
