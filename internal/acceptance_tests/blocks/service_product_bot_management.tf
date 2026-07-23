resource "fastly_service_product_bot_management" "test" {
  service_id   = {{.SERVICE_ID_REF}}
  contentguard = "{{.CONTENT_GUARD}}"
}
