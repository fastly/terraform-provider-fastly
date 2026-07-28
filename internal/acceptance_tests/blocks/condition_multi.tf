  condition {
    name      = "{{.CONDITION_NAME_1}}"
    type      = "REQUEST"
    statement = "req.url ~ \"^/admin\""
  }

  condition {
    name      = "{{.CONDITION_NAME_2}}"
    type      = "CACHE"
    statement = "beresp.status == 200"
    priority  = 20
  }
