  gzip {
    name          = "{{.GZIP_NAME}}"
    content_types = ["text/html", "text/css"]
    extensions    = ["css", "js"]
  }
