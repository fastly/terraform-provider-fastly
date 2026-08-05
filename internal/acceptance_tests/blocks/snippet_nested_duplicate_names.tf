snippet {
  name     = "duplicate"
  type     = "recv"
  priority = 100
  content  = file("{{.SNIPPET_FILE_PATH_ONE}}")
}

snippet {
  name     = "duplicate"
  type     = "deliver"
  priority = 50
  content  = file("{{.SNIPPET_FILE_PATH_TWO}}")
}
