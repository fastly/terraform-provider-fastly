#...

resource "fastly_service_dynamic_snippet_content" "my_dyn_content" {
  for_each = {
    for d in fastly_service_vcl.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id       = fastly_service_vcl.myservice.id
  snippet_id       = each.value.snippet_id
  manage_snippets = true
  content          = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

  lifecycle {
    ignore_changes = [content, ]
  }
}