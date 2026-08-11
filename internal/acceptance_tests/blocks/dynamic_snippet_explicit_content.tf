resource "fastly_service_dynamic_snippet_content" "test" {
  service_id = fastly_service_cdn.test.id
  snippet_id = fastly_service_dynamic_vcl_snippet.test.snippet_id

  content         = {{.DYNAMIC_SNIPPET_INLINE_CONTENT}}
  manage_snippets = {{.MANAGE_SNIPPETS}}
}
