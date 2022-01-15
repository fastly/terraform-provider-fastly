locals {
  domains = [
    "example.com",
    "*.example.com",
  ]
  aws_route53_zone_id = "your_route53_zone_id"
}

resource "fastly_service_v1" "example" {
  name = "example-service"

  dynamic domain {
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

resource "fastly_tls_subscription" "example" {
  domains               = [for domain in fastly_service_v1.example.domain : domain.name]
  certificate_authority = "lets-encrypt"
}

# Set up DNS record for managed DNS domain validation method
resource "aws_route53_record" "domain_validation" {
  depends_on = [fastly_tls_subscription.example]

  # NOTE: in this example, two domains are added to the cert ("example.com" and "*.example.com").
  # The "managed_dns_challenges" read-only attribute only includes one object
  # for "_acme-challenge.example.com" as the challenge record is common in these two domains.
  # 
  # In order to process a cert contains wildcard entries, remove wildcard prefix "*." from the key
  # and use ellipsis (...) to group results by key to avoid "Duplicate object key" error.
  # Therefore, a key may have multiple elements. For example, domains "example.com" and "*.example.com"
  # find the exact same object in the "managed_dns_challenges" attribute due to the "if" statement below.
  # 
  # A simplified version of this complex "for_each" usage would be:
  # ```
  # for_each = {
  #   for challenge in fastly_tls_subscription.example.managed_dns_challenges :
  #   trimprefix(challenge.record_name, "_acme-challenge.") => challenge
  # }
  # ```
  # but since the "managed_dns_challenges" attribute is only known after apply,
  # you will need to create this resource separately ("--target" option) and may not be ideal.
  for_each = {
    for domain in fastly_tls_subscription.example.domains :
    replace(domain, "*.", "") => element([
      for obj in fastly_tls_subscription.example.managed_dns_challenges :
      obj if obj.record_name == "_acme-challenge.${replace(domain, "*.", "")}"
    ], 0)...
  }
  
  # only reads the first element in the list since all elements are exactly the same (see above)
  name            = each.value[0].record_name
  type            = each.value[0].record_type
  zone_id         = local.aws_route53_zone_id
  allow_overwrite = true
  records         = [each.value[0].record_value]
  ttl             = 60
}

# Resource that other resources can depend on if they require the certificate to be issued
resource "fastly_tls_subscription_validation" "example" {
  subscription_id = fastly_tls_subscription.example.id
  depends_on      = [aws_route53_record.domain_validation]
}