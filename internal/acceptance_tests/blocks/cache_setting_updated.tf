  cache_setting {
    name      = "{{.CACHE_SETTING_NAME}}"
    action    = "pass"
    ttl       = 7200
    stale_ttl = 300
  }
