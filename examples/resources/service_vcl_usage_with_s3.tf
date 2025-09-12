resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address       = "http-me.fastly.dev"
    name          = "Glitch Test Site"
    port          = 80
    override_host = "http-me.fastly.dev"
  }

  header {
    destination = "http.x-amz-request-id"
    type        = "cache"
    action      = "delete"
    name        = "remove x-amz-request-id"
  }

  gzip {
    name          = "file extensions and content types"
    extensions    = ["css", "js"]
    content_types = ["text/html", "text/css"]
  }

  force_destroy = true
}

resource "aws_s3_bucket" "website" {
  bucket = "demo.notexample.com"
}
