  condition {
    name      = "{{.CONDITION_NAME}}"
    type      = "CACHE"
    statement = "beresp.status == 200"
  }
