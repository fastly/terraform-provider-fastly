resource "fastly_kvstore" "kv1" {
  name = "{{.KVSTORE_NAME_1}}"
}

resource "fastly_kvstore" "kv2" {
  name = "{{.KVSTORE_NAME_2}}"
}
