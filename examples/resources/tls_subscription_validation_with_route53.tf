locals {
  domain_name = "example.com"
}

resource "fastly_service_vcl" "example" {
  name = "example-service"

  domain {
    name = local.domain_name
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

data "aws_route53_zone" "demo" {
  name         = local.domain_name
  private_zone = false
}

# Set up DNS record for managed DNS domain validation method
resource "aws_route53_record" "domain_validation" {
  name            = fastly_tls_subscription.example.managed_dns_challenge.record_name
  type            = fastly_tls_subscription.example.managed_dns_challenge.record_type
  zone_id         = data.aws_route53_zone.demo.id
  allow_overwrite = true
  records         = [fastly_tls_subscription.example.managed_dns_challenge.record_value]
  ttl             = 60
}

# Resource that other resources can depend on if they require the certificate to be issued
resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [aws_route53_record.domain_validation]
}