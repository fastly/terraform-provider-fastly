resource "fastly_service_vcl" "myservice" {
  name = "snippet_test"

  domain {
    name    = "snippet.fastlytestdomain.com"
    comment = "snippet test"
  }

  backend {
    address = "http-me.glitch.me"
    name    = "Glitch Test Site"
    port    = 80
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet One"
    type     = "recv"
    priority = 110
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet Two"
    type     = "recv"
    priority = 110
  }

  default_host = "http-me.glitch.me"

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content" "my_dyn_content_one" {
  for_each = {
  for d in fastly_service_vcl.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet One"
  }

  service_id = fastly_service_vcl.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header-one = \"true\";\n}"

}

resource "fastly_service_dynamic_snippet_content" "my_dyn_content_two" {
  for_each = {
  for d in fastly_service_vcl.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet Two"
  }

  service_id = fastly_service_vcl.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header-two = \"true\";\n}"

}