resource "fastly_kvstore" "store" {
  name = "{{.KVSTORE_NAME}}"
}
