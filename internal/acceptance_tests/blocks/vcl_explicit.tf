resource "fastly_service_vcl" "main" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "{{.VCL_NAME}}"
  main       = true
  content    = file("{{.VCL_FILE_PATH}}")
}
