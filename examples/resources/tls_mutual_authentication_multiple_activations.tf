terraform {
  required_providers {
    dnsimple = {
      source  = "dnsimple/dnsimple"
      version = "1.5.0"
    }
    fastly = {
      source  = "fastly/fastly"
      version = "5.7.2"
    }
  }
}

variable "dnsimple_token" {
  type = string
}

variable "dnsimple_account" {
  type = string
}

provider "dnsimple" {
  account = var.dnsimple_account
  token   = var.dnsimple_token
}

variable "zone" {
  type    = string
  default = "example.com"
}

resource "fastly_service_vcl" "example" {
  name = "example"
  domain {
    name = "foo.${var.zone}"
  }
  domain {
    name = "bar.${var.zone}"
  }
  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }
  force_destroy = true
}

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name]
  certificate_authority = "certainly"
}

resource "dnsimple_zone_record" "example_acme_challenge" {
  for_each = {
    for domain in fastly_tls_subscription.example.domains : domain => one([
      for obj in fastly_tls_subscription.example.managed_dns_challenges : obj if obj.record_name == "_acme-challenge.${domain}"
    ])
  }
  name      = each.value.record_name
  ttl       = "60"
  type      = each.value.record_type
  value     = each.value.record_value
  zone_name = var.zone
}

resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [dnsimple_zone_record.example_acme_challenge]
}

data "fastly_tls_configuration" "default" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.example]
}

resource "dnsimple_zone_record" "foo" {
  name      = "foo"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

resource "dnsimple_zone_record" "bar" {
  name      = "bar"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

# IMPORTANT: The subscription's certificate_id attribute is initially empty.
# So we can't reference the certificate_id attribute directly (not until a state refresh).
# This means we need to use the subscription data source instead.
# We need this data source to wait for the subscription process to complete.
# Once complete we'll have a Certificate ID we can reference as input to the `fastly_tls_activation_ids` data source.
data "fastly_tls_subscription" "example" {
  id         = fastly_tls_subscription.example.id
  depends_on = [fastly_tls_subscription_validation.example]
}

data "fastly_tls_activation_ids" "example" {
  certificate_id = one(data.fastly_tls_subscription.example.certificate_ids)
}

resource "fastly_tls_mutual_authentication" "example" {
  activation_ids = data.fastly_tls_activation_ids.example.ids
  cert_bundle    = "-----BEGIN CERTIFICATE-----\n<REDACTED>\n-----END CERTIFICATE-----"
  enforced       = true
}
