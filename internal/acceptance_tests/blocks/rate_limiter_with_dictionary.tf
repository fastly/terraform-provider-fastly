  dictionary {
    name = "{{.DICTIONARY_NAME}}"
  }

  rate_limiter {
    name                 = "{{.RATE_LIMITER_NAME}}"
    action               = "log_only"
    client_key           = ["req.http.Fastly-Client-IP"]
    http_methods         = ["GET"]
    logger_type          = "s3"
    penalty_box_duration = 5
    rps_limit            = 50
    uri_dictionary_name  = "{{.DICTIONARY_NAME}}"
    window_size          = 60
  }
