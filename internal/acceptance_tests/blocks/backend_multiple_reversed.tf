backend {
  name              = "a"
  address           = "a.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "a.example.com"
  ssl_sni_hostname  = "a.example.com"
  weight            = 100
}

backend {
  name              = "b"
  address           = "b.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "b.example.com"
  ssl_sni_hostname  = "b.example.com"
  weight            = 100
}
