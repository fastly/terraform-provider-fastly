resource "fastly_kvstore" "store_1" {
  name = "{{.KVSTORE_NAME_1}}"
}

resource "fastly_kvstore" "store_2" {
  name = "{{.KVSTORE_NAME_2}}"
}

resource "fastly_kvstore" "store_3" {
  name = "{{.KVSTORE_NAME_3}}"
}

data "fastly_kvstores" "example" {
  depends_on = [
    fastly_kvstore.store_1,
    fastly_kvstore.store_2,
    fastly_kvstore.store_3,
  ]
}
