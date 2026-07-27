resource "fastly_service_product_fanout" "test" {
  service_id = {{.SERVICE_ID_REF}}
}
