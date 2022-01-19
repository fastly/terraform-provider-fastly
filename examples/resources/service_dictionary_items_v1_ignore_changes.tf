#...

resource "fastly_service_dictionary_items_v1" "items" {
  for_each      = {
  for d in fastly_service_v1.myservice.dictionary : d.name => d if d.name == var.mydict_name
  }
  service_id    = fastly_service_v1.myservice.id
  dictionary_id = each.value.dictionary_id

  items = {
    key1 : "value1"
    key2 : "value2"
  }

  lifecycle {
    ignore_changes = [items,]
  }
}