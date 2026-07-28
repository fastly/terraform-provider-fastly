vcl {
  name    = "include_only"
  main    = false
  content = file("{{.VCL_INCLUDE_FILE_PATH}}")
}
