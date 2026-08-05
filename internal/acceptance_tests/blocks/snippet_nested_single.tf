snippet {
  name     = "{{.SNIPPET_NAME}}"
  type     = "{{.SNIPPET_TYPE}}"
  priority = {{.SNIPPET_PRIORITY}}
  content  = file("{{.SNIPPET_FILE_PATH}}")
}
