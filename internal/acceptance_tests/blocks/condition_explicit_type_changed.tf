resource "fastly_service_condition" "test" {
  service_id = fastly_service_cdn.test.id
  version    = {{.SERVICE_VERSION}}
  name       = "{{.CONDITION_NAME}}"
  type       = "CACHE"
  statement  = "beresp.status == 200"
}
