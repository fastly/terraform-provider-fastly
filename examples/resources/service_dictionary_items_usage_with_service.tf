variable "mydict_name" {
  type    = string
  default = "My Dictionary"
}

resource "fastly_service_vcl" "myservice" {
  #...
  dictionary {
    name = var.mydict_name
  }
  #...
}

resource "fastly_service_dictionary_items" "items" {
  service_id    = fastly_service_vcl.myservice.id
  dictionary_id = {for s in fastly_service_vcl.myservice.dictionary : s.name => s.dictionary_id}[var.mydict_name]

  items = {
    key1 : "value1"
    key2 : "value2"
  }
}