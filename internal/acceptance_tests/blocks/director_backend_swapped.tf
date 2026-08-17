backend {
  name              = "{{.BACKEND_NAME_2}}"
  address           = "api2.example.com"
  port              = 443
  use_ssl           = true
  ssl_cert_hostname = "api2.example.com"
  ssl_sni_hostname  = "api2.example.com"
}

director {
  name     = "{{.DIRECTOR_NAME}}"
  backends = ["{{.BACKEND_NAME_2}}"]
}
