  gzip {
    name          = "{{.GZIP_NAME_1}}"
    content_types = ["text/html"]
    extensions    = ["css"]
  }

  gzip {
    name          = "{{.GZIP_NAME_2}}"
    extensions    = ["js"]
  }
