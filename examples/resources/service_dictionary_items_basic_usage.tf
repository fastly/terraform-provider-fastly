variable "mydict_name" {
  type = string
  default = "My Dictionary"
}

resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "http-me.glitch.me"
    name    = "Glitch Test Site"
    port    = 80
  }

  dictionary {
    name       = var.mydict_name
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items" "items" {
  for_each = {
  for d in fastly_service_vcl.myservice.dictionary : d.name => d if d.name == var.mydict_name
  }
  service_id = fastly_service_vcl.myservice.id
  dictionary_id = each.value.dictionary_id

  items = {
    key1: "value1"
    key2: "value2"
  }
}