resource "fastly_service_condition" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.CONDITION_NAME}}"
  type       = "REQUEST"
  statement  = "req.url ~ \"^/admin\""
}

resource "fastly_service_backend" "origin" {
  service_id         = fastly_service_cdn.test.id
  version            = {{.SERVICE_VERSION}}
  name               = "{{.BACKEND_NAME}}"
  address            = "api.example.com"
  port               = 443
  use_ssl            = true
  request_condition  = fastly_service_condition.test.name
}
