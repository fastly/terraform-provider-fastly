resource "fastly_service_vcl" "example" {
  name = "example_service"

  domain {
    name    = "example.com"
    comment = "example domain"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
    port    = 80
  }

  force_destroy = true
}

resource "fastly_acl" "example" {
  name       = "example_acl"
  service_id = fastly_service_vcl.example.id
}

# Optionally manage ACL entries
resource "fastly_service_acl_entries" "entries" {
  service_id = fastly_service_vcl.example.id
  acl_id     = fastly_acl.example.acl_id

  entry {
    ip      = "192.168.0.1"
    subnet  = "24"
    comment = "Block internal IPs"
    negated = true
  }
  
  entry {
    ip      = "10.0.0.1"
    subnet  = "16"
    comment = "Allow office IPs"
  }
}