resource "fastly_ngwaf_workspace" "demo" {
  name = "demofastly"

  description = "A reference setup"
  mode        = "block"
  attack_signal_thresholds {
    immediate = true
  }
  default_blocking_response_code = 406
  ip_anonymization               = null
  client_ip_headers              = []
}
