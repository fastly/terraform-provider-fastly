resource "fastly_service_cdn" "test" {
  name          = "{{.SERVICE_NAME}}"
  comment       = "{{.SERVICE_COMMENT}}"
  force_destroy = true
}
{{.RESOURCES}}
