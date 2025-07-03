data "fastly_ngwaf_workspaces" "ngwaf_workspaces" {}

output "fastly_ngwaf_workspaces_all" {
  value = data.fastly_ngwaf_workspaces.workspaces
}

output "fastly_ngwaf_workspaces_filtered" {
  # get the workspace with the name "Example Workspace"
  value = one([for workspace in data.fastly_ngwaf_workspaces.workspaces.details : workspace.id if workspace.name == "Example Workspace"])
}
