data "fastly_tls_configuration_ids" "example" {}

resource "fastly_tls_activation" "example" {
  configuration_id = data.fastly_tls_configuration_ids.example.ids[0]
  // ...
}
