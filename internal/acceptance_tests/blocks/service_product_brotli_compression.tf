resource "fastly_service_product_brotli_compression" "test" {
  service_id = {{.SERVICE_ID_REF}}
}
