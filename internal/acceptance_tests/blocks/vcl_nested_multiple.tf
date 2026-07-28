vcl {
  name    = "{{.VCL_MAIN_NAME}}"
  main    = true
  content = file("{{.VCL_MAIN_FILE_PATH}}")
}

vcl {
  name    = "{{.VCL_INCLUDE_NAME}}"
  main    = false
  content = file("{{.VCL_INCLUDE_FILE_PATH}}")
}
