resource "fastly_service_dynamic_snippet_content" "test" {
  service_id = fastly_service_cdn_auto.test.id
  snippet_id = {
    for s in fastly_service_cdn_auto.test.dynamic_snippet : s.name => s.snippet_id
  }["{{.DYNAMIC_SNIPPET_NAME}}"]

  content         = {{.DYNAMIC_SNIPPET_INLINE_CONTENT}}
  manage_snippets = {{.MANAGE_SNIPPETS}}
}
