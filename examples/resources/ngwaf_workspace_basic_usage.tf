resource "fastly_ngwaf_workspace" "demo" {
  name = "demofastly"

  force_destroy = true
}