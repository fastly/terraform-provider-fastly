data "fastly_tls_configuration" "example" {
  default = true
}

resource "fastly_tls_activation" "example" {
  configuration_id = data.fastly_tls_configuration.example.id
  // ...
}
