  condition {
    name      = "{{.CONDITION_NAME}}"
    type      = "CACHE"
    statement = "req.url ~ \"\\.(css|js|html)$\""
  }
