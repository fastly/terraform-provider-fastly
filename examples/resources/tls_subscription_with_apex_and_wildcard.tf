terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.55.0"
    }
    fastly = {
      source  = "fastly/fastly"
      version = "3.1.0"
    }
  }
}

# NOTE: Creating a hosted zone will automatically create SOA/NS records.
resource "aws_route53_zone" "production" {
  name = "example.com"
}

resource "aws_route53domains_registered_domain" "example" {
  domain_name = "example.com"

  dynamic "name_server" {
    for_each = aws_route53_zone.production.name_servers

    content {
      name = name_server.value
    }
  }
}

locals {
  domains = [
    "example.com",
    "*.example.com",
  ]
}

resource "fastly_service_vcl" "example" {
  name = "example-service"

  dynamic "domain" {
    for_each = local.domains

    content {
      name = domain.value
    }
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_subscription" "testing_tls" {
  domains               = [for domain in fastly_service_vcl.testing-tls.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

resource "aws_route53_record" "domain_validation" {
  depends_on = [fastly_tls_subscription.testing_tls]

  for_each = {
    # The following `for` expression (due to the outer {}) will produce an object with key/value pairs.
    # In this example we are defining an apex (example.com) and a wildcard (*.example.com) which causes the API to return a single challenge (e.g. _acme-challenge.example.com)
    # To ensure we can match the single challenge for both domains we need to normalise the wildcard domain.
    # The 'key' is the normalised domain name (e.g. example.com)
    # The 'value' is the single 'challenge' object whose record_name matches the normalised version of the domain (e.g. record_name is _acme-challenge.example.com).
    for domain in fastly_tls_subscription.testing_tls.domains :
    replace(domain, "*.", "") => element([
      for obj in fastly_tls_subscription.testing_tls.managed_dns_challenges :
      obj if obj.record_name == "_acme-challenge.${replace(domain, "*.", "")}" # We use an `if` conditional to filter the list to a single element
    ], 0)...                                                                   # `element()` returns the first object in the list which should be the relevant 'challenge' object we need
    # The ellipsis ... avoids Terraform complaining that the resulting object will contain multiple keys that are duplicates (e.g. multiple 'example.com' keys).
    # It essentially groups the 'values' (the single challenge) under the common key (the normalised domain).
    # Then below we extract the first value (as they'll all be the same 'challenge' value).
  }

  name            = each.value[0].record_name
  type            = each.value[0].record_type
  zone_id         = aws_route53_zone.production.zone_id
  allow_overwrite = true
  records         = [each.value[0].record_value]
  ttl             = 60
}

# This is a resource that other resources can depend on if they require the certificate to be issued.
# NOTE: Internally the resource keeps retrying `GetTLSSubscription` until no error is returned (or the configured timeout is reached).
resource "fastly_tls_subscription_validation" "testing_tls" {
  subscription_id = fastly_tls_subscription.testing_tls.id
  depends_on      = [aws_route53_record.domain_validation]
}

# This data source lists all configuration and uses the `default` attribute to narrow down the configuration to just one object.
# If the filtered list has a length that is not exactly one element, you'll see an error returned.
# That single TLS configuration element is then returned.
data "fastly_tls_configuration" "default_tls" {
  default    = true
  depends_on = [fastly_tls_subscription_validation.testing_tls]
}

# Once validation is complete and we've retrieved the TLS configuration data, we can create multiple records...

resource "aws_route53_record" "apex" {
  name    = "example.com"
  records = [for record in data.fastly_tls_configuration.default_tls.dns_records : record.record_value if record.record_type == "A"]
  ttl     = 300
  type    = "A"
  zone_id = aws_route53_zone.production.zone_id
}

# NOTE: This subdomain matches our Fastly service because of the wildcard domain (`*.example.com`) that was added to the service.
resource "aws_route53_record" "subdomain" {
  name    = "test.example.com"
  records = [for record in data.fastly_tls_configuration.default_tls.dns_records : record.record_value if record.record_type == "CNAME"]
  ttl     = 300
  type    = "CNAME"
  zone_id = aws_route53_zone.production.zone_id
}
