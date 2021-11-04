resource "fastly_service_v1" "myservice" {
  name = "snippet_test"

  domain {
    name    = "snippet.fastlytestdomain.com"
    comment = "snippet test"
  }

  backend {
    address = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dynamicsnippet {
    name     = "My Dynamic Snippet"
    type     = "recv"
    priority = 110
  }

  default_host = "tftesting.tftesting.net.s3-website-us-west-2.amazonaws.com"

  force_destroy = true
}

resource "fastly_service_dynamic_snippet_content_v1" "my_dyn_content" {
  for_each = {
    for d in fastly_service_v1.myservice.dynamicsnippet : d.name => d if d.name == "My Dynamic Snippet"
  }
  service_id = fastly_service_v1.myservice.id
  snippet_id = each.value.snippet_id

  content = "if ( req.url ) {\n set req.http.my-snippet-test-header = \"true\";\n}"

}