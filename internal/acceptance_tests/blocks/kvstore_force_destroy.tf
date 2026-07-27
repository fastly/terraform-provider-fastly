resource "fastly_kvstore" "store" {
  name          = "{{.KVSTORE_NAME}}"
  force_destroy = true
}
