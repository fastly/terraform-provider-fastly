  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME}}"
    action               = "response"
    client_key           = ["req.http.Fastly-Client-IP"]
    http_methods         = ["GET", "POST"]
    penalty_box_duration = 10
    rps_limit            = 100
    window_size          = 10

    response = {
      content      = "Rate limit exceeded"
      content_type = "text/plain"
      status       = 429
    }
  }
