#...

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content" {
  for_each   = {
  for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

  lifecycle {
    ignore_changes = [content,]
  }
}