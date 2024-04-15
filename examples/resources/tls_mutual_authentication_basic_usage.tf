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
    name = "www.${var.zone}"
  }
  backend {
    address = "httpbin.org"
    name    = "httpbin"
  }
  force_destroy = true
}

resource "fastly_tls_subscription" "www" {
  domains               = [for domain in fastly_service_vcl.example.domain : domain.name if domain.name == "www.${var.zone}"]
  certificate_authority = "certainly"
}

resource "dnsimple_zone_record" "www_acme_challenge" {
  name      = "_acme-challenge.www"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for obj in fastly_tls_subscription.www.managed_dns_challenges : obj.record_value if obj.record_name == "_acme-challenge.www.${var.zone}"])
  zone_name = var.zone
}

resource "fastly_tls_subscription_validation" "www" {
  subscription_id = fastly_tls_subscription.www.id
  depends_on      = [dnsimple_zone_record.www_acme_challenge]
}

data "fastly_tls_configuration" "default" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.www]
}

resource "dnsimple_zone_record" "www" {
  name      = "www"
  ttl       = "60"
  type      = "CNAME"
  value     = one([for record in data.fastly_tls_configuration.default.dns_records : record.record_value if record.record_type == "CNAME"])
  zone_name = var.zone
}

data "fastly_tls_activation" "www" {
  domain     = "www.example.com"
  depends_on = [dnsimple_zone_record.www]
}

resource "fastly_tls_mutual_authentication" "www" {
  activation_ids = [data.fastly_tls_activation.www.id]
  cert_bundle    = "-----BEGIN CERTIFICATE-----\n<REDACTED>\n-----END CERTIFICATE-----"
  enforced       = true
}
