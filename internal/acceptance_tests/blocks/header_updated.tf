  header {
    name          = "{{.HEADER_NAME}}"
    action        = "set"
    type          = "request"
    destination   = "http.X-Custom"
    source        = "req.http.Host"
    priority      = 10
    ignore_if_set = true
  }
