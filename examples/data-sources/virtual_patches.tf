data "fastly_ngwaf_virtual_patches" "list_patches" {
    workspace_id = fastly_ngwaf_workspace.test_virtual_patches_workspace.id
}

output "fastly_ngwaf_virtual_patches_all" {
  value = data.fastly_ngwaf_virtual_patches.list_patches
}