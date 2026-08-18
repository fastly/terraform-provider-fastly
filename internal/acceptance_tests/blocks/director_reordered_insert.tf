backend {
  name              = "{{.BACKEND_NAME_C}}"
  address           = "api3.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "api3.example.com"
  ssl_sni_hostname  = "api3.example.com"
}

backend {
  name              = "{{.BACKEND_NAME_A}}"
  address           = "api.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "api.example.com"
  ssl_sni_hostname  = "api.example.com"
}

backend {
  name              = "{{.BACKEND_NAME_B}}"
  address           = "api2.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "api2.example.com"
  ssl_sni_hostname  = "api2.example.com"
}

director {
  name     = "{{.DIRECTOR_NAME_C}}"
  backends = ["{{.BACKEND_NAME_C}}"]
}

director {
  name     = "{{.DIRECTOR_NAME_A}}"
  backends = ["{{.BACKEND_NAME_A}}"]
}

director {
  name     = "{{.DIRECTOR_NAME_B}}"
  backends = ["{{.BACKEND_NAME_B}}"]
}
