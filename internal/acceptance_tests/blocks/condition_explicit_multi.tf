resource "fastly_service_condition" "test1" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.CONDITION_NAME_1}}"
  type       = "REQUEST"
  statement  = "req.url ~ \"^/admin\""
}

resource "fastly_service_condition" "test2" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.CONDITION_NAME_2}}"
  type       = "CACHE"
  statement  = "beresp.status == 200"
}
