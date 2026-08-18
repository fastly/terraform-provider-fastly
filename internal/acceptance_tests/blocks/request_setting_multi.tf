  condition {
    name      = "{{.CONDITION_NAME_1}}"
    type      = "REQUEST"
    statement = "req.url ~ \"^/admin\""
  }

  condition {
    name      = "{{.CONDITION_NAME_2}}"
    type      = "REQUEST"
    statement = "req.url ~ \"^/api\""
  }

  request_setting {
    name              = "{{.REQUEST_SETTING_NAME_1}}"
    request_condition = "{{.CONDITION_NAME_1}}"
    max_stale_age     = 120
  }

  request_setting {
    name              = "{{.REQUEST_SETTING_NAME_2}}"
    request_condition = "{{.CONDITION_NAME_2}}"
    action            = "pass"
  }
