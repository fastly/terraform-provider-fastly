  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME}}"
    action               = "log_only"
    client_key           = ["req.http.Fastly-Client-IP", "req.http.User-Agent"]
    http_methods         = ["GET", "PUT"]
    logger_type          = "bigquery"
    penalty_box_duration = 15
    rps_limit            = 75
    window_size          = 1
  }
