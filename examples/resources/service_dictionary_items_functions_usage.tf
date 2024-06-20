// Local variables used when formatting values for the "My Project Dictionary" example
locals {
  dictionary_name = "My Project Dictionary"
  host_base       = "demo.ocnotexample.com"
  host_divisions  = ["alpha", "beta", "gamma", "delta"]
}

// Define the standard service that will be used to manage the dictionaries.
resource "fastly_service_vcl" "myservice" {
  name = "demofastly"

  domain {
    name    = "demo.ocnotexample.com"
    comment = "demo"
  }

  backend {
    address = "http-me.glitch.me"
    name    = "Glitch Test Site"
    port    = 80
  }

  dictionary {
    name = local.dictionary_name
  }

  force_destroy = true
}

// This resource is dynamically creating the items from the local variables through for expressions and functions.
resource "fastly_service_dictionary_items" "project" {
  for_each      = {
  for d in fastly_service_vcl.myservice.dictionary : d.name => d if d.name == local.dictionary_name
  }
  service_id    = fastly_service_vcl.myservice.id
  dictionary_id = each.value.dictionary_id
  items         = {
  for division in local.host_divisions :
  division => format("%s.%s", division, local.host_base)
  }
}