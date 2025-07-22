resource "fastly_ngwaf_redaction" "demo_redaction" {
  field                        = "some field"
  type                         = "request_header"
  workspace_id                 = fastly_ngwaf_workspace.demo.id
}
