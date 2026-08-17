  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME_1}}"
    action               = "log_only"
    client_key           = ["req.http.Fastly-Client-IP"]
    http_methods         = ["GET"]
    logger_type          = "s3"
    penalty_box_duration = 5
    rps_limit            = 50
    window_size          = 60
  }

  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME_2}}"
    action               = "response"
    client_key           = ["req.http.Fastly-Client-IP"]
    http_methods         = ["POST", "DELETE"]
    penalty_box_duration = 10
    rps_limit            = 100
    window_size          = 10

    response = {
      content      = "Rate limit exceeded"
      content_type = "text/plain"
      status       = 429
    }
  }
