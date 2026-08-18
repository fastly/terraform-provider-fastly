  header {
    name              = "{{.HEADER_NAME}}"
    action            = "delete"
    type              = "request"
    destination       = "http.x-amz-request-id"
    request_condition = "{{.CONDITION_NAME}}"
  }
