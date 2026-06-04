  backend {
    name              = "backend-primary"
    address           = "api-primary.example.com"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "api-primary.example.com"
    ssl_sni_hostname  = "api-primary.example.com"
    weight            = 100
  }

  backend {
    name              = "backend-secondary"
    address           = "api-secondary.example.com"
    port              = 443
    use_ssl           = true
    ssl_cert_hostname = "api-secondary.example.com"
    ssl_sni_hostname  = "api-secondary.example.com"
    weight            = 50
  }
