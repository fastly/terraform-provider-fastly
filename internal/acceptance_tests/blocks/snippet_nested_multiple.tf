snippet {
  name     = "{{.SNIPPET_NAME_ONE}}"
  type     = "recv"
  priority = 100
  content  = file("{{.SNIPPET_FILE_PATH_ONE}}")
}

snippet {
  name     = "{{.SNIPPET_NAME_TWO}}"
  type     = "deliver"
  priority = 50
  content  = file("{{.SNIPPET_FILE_PATH_TWO}}")
}
