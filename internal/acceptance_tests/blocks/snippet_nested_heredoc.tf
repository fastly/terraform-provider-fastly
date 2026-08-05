snippet {
  name     = "{{.SNIPPET_NAME}}"
  type     = "{{.SNIPPET_TYPE}}"
  priority = {{.SNIPPET_PRIORITY}}
  content  = <<SNIPPET
{{.SNIPPET_HEREDOC_CONTENT}}
SNIPPET
}
