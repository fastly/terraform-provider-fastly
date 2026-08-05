resource "fastly_service_vcl_snippet" "test" {
  service_id = fastly_service_cdn.test.id
  version    = 1
  name       = "{{.SNIPPET_NAME}}"
  type       = "{{.SNIPPET_TYPE}}"
  priority   = {{.SNIPPET_PRIORITY}}
  content    = {{.SNIPPET_INLINE_CONTENT}}
}
