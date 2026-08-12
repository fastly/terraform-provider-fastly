resource "fastly_service_dynamic_vcl_snippet" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "{{.DYNAMIC_SNIPPET_NAME}}"
  type       = "{{.DYNAMIC_SNIPPET_TYPE}}"
  priority   = {{.DYNAMIC_SNIPPET_PRIORITY}}
}
