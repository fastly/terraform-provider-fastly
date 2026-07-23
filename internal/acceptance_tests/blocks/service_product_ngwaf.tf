resource "fastly_service_product_ngwaf" "test" {
  service_id   = {{.SERVICE_ID_REF}}
  workspace_id = "{{.NGWAF_WORKSPACE_ID}}"
{{if .NGWAF_TRAFFIC_RAMP}}  traffic_ramp = {{.NGWAF_TRAFFIC_RAMP}}
{{end}}}
