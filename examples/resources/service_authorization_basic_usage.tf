resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }
}


resource "fastly_user" "user" {
  login = "demo@example.com"
  name  = "Demo User"
}

resource "fastly_service_authorization" "auth" {
  service_id = fastly_service_vcl.demo.id
  user_id    = fastly_user.user.id
  permission = "purge_all"
}
