resource "fastly_service_vcl" "myservice" {
  #...
  dynamicsnippet {
    name     = "My Dynamic Snippet"
    type     = "recv"
    priority = 110
  }
  #...
}

resource "fastly_service_dynamic_snippet_content" "my_dyn_content" {
  service_id = fastly_service_vcl.myservice.id
  snippet_id = {for s in fastly_service_vcl.myservice.dynamicsnippet : s.name => s.snippet_id}["My Dynamic Snippet"]

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

}