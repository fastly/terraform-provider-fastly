  cache_setting {
    name      = "{{.CACHE_SETTING_NAME}}"
    action    = "cache"
    ttl       = 3600
    stale_ttl = 120
  }
