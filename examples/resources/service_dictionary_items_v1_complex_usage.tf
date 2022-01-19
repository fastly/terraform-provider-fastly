variable "mydict" {
  type    = object({ name = string, items = map(string) })
  default = {
    name  = "My Dictionary"
    items = {
      key1 : "value1x"
      key2 : "value2x"
    }
  }
}

resource "fastly_service_v1" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "demo.notexample.com.s3-website-us-west-2.amazonaws.com"
    name    = "AWS S3 hosting"
    port    = 80
  }

  dictionary {
    name = var.mydict.name
  }

  force_destroy = true
}

resource "fastly_service_dictionary_items_v1" "items" {
  for_each      = {
  for d in fastly_service_v1.myservice.dictionary : d.name => d if d.name == var.mydict.name
  }
  service_id    = fastly_service_v1.myservice.id
  dictionary_id = each.value.dictionary_id
  items         = var.mydict.items
}