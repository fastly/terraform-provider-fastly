resource "fastly_service_vcl" "demo" {
  #...
}


resource "fastly_user" "user" {
  # ...
}

resource "fastly_service_authorization" "auth" {
  service_id = fastly_service_vcl.demo.id
  user_id    = fastly_user.user.id
  permission = "purge_all"
}
