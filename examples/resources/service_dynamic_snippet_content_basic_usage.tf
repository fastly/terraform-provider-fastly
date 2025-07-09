resource "fastly_service_vcl" "myservice" {
  name = "snippet_test"

  domain {
    name    = "snippet.fastlytestdomain.com"
    comment = "snippet test"
  }

  backend {
    address = "http-me.fastly.dev"
    name    = "Glitch Test Site"
    port    = 80
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet"
    type     = "recv"
    priority = 110
  }

  default_host = "http-me.fastly.dev"

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content" "my_dyn_content" {
  for_each = {
    for d in fastly_service_vcl.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id = fastly_service_vcl.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

}