resource "fastly_ngwaf_virtual_patches" "demo" {
  action            = "block"
  enabled           = true
  virtual_patch_id    = "CVE-2017-5638"
  workspace_id       = fastly_ngwaf_workspace.demo.id
}