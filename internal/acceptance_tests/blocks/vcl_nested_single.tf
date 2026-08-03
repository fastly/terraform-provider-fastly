vcl {
  name    = "{{.VCL_NAME}}"
  main    = true
  content = file("{{.VCL_FILE_PATH}}")
}
