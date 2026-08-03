  condition {
    name      = "{{.CONDITION_NAME}}"
    type      = "REQUEST"
    statement = "req.url ~ \"^/private\""
    priority  = 5
  }
