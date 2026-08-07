vcl {
  name    = "{{.VCL_NAME}}"
  main    = true
  content = <<-VCL
{{.VCL_HEREDOC_CONTENT}}
VCL
}
