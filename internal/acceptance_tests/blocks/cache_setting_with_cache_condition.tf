  cache_setting {
    name            = "{{.CACHE_SETTING_NAME}}"
    action          = "cache"
    ttl             = 3600
    cache_condition = "{{.CONDITION_NAME}}"
  }
