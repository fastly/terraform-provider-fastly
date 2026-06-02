resource "fastly_service_compute_auto" "test" {
  name          = "SERVICE_NAME"
  comment       = "Updated by Terraform"
  force_destroy = true
}
