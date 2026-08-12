snippet {
  name     = "{{.SNIPPET_NAME}}"
  type     = "recv"
  priority = 100
  content  = file("{{.SNIPPET_FILE_PATH}}")
}

dynamic_snippet {
  name     = "{{.SNIPPET_NAME}}"
  type     = "recv"
  priority = 100
}
