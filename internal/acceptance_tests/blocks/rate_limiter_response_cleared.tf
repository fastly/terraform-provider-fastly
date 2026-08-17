  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME}}"
    action               = "log_only"
    logger_type          = "s3"
    client_key           = ["req.http.Fastly-Client-IP"]
    http_methods         = ["GET", "POST"]
    penalty_box_duration = 10
    rps_limit            = 100
    window_size          = 10
  }
