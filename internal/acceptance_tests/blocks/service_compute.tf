resource "fastly_service_compute" "test" {
  name          = "{{.SERVICE_NAME}}"
  comment       = "{{.SERVICE_COMMENT}}"
  force_destroy = true
}
{{.RESOURCES}}
