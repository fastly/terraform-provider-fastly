locals {
  domain_name = "example.com"
}

resource "fastly_service_v1" "example" {
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
  domains               = [for domain in fastly_service_v1.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

data "aws_route53_zone" "demo" {
  name         = local.domain_name
  private_zone = false
}

# Set up DNS record for managed DNS domain validation method
resource "aws_route53_record" "domain_validation" {
  depends_on = [fastly_tls_subscription.example]

  for_each        = {
  for domain in fastly_tls_subscription.example.domains :
  domain => element([
  for obj in fastly_tls_subscription.example.managed_dns_challenges :
  obj if obj.record_name == "_acme-challenge.${domain}"
  ], 0)
  }
  name            = each.value.record_name
  type            = each.value.record_type
  zone_id         = data.aws_route53_zone.demo.id
  allow_overwrite = true
  records         = [each.value.record_value]
  ttl             = 60
}

# Resource that other resources can depend on if they require the certificate to be issued
resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [aws_route53_record.domain_validation]
}