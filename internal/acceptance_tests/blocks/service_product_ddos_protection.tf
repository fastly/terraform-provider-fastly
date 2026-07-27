resource "fastly_service_product_ddos_protection" "test" {
  service_id = {{.SERVICE_ID_REF}}
  mode       = "{{.DDOS_MODE}}"
}
