resource "fastly_service_cdn_auto" "{{.LABEL}}" {
  name          = "{{.SERVICE_NAME}}"
  force_destroy = true

  domain {
    name = "{{.DOMAIN_NAME}}"
  }
}
