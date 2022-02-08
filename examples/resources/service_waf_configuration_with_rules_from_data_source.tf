variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}

resource "fastly_service_vcl" "demo" {
  name = "demofastly"

  domain {
    name    = "example.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "WAF_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # This condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic.
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name              = "WAF_Response"
    status            = "403"
    response          = "Forbidden"
    content_type      = "text/html"
    content           = "<html><body>Forbidden</body></html>"
    request_condition = "WAF_always_false"
  }
  waf {
    prefetch_condition = "WAF_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

data "fastly_waf_rules" "owasp" {
  publishers = ["owasp"]
}

resource "fastly_service_waf_configuration" "waf" {
  waf_id                          = fastly_service_vcl.demo.waf[0].waf_id
  http_violation_score_threshold  = 100

  dynamic "rule" {
    for_each = data.fastly_waf_rules.owasp.rules
    content {
      modsec_rule_id = rule.value.modsec_rule_id
      revision       = rule.value.latest_revision_number
      status         = lookup(var.type_status, rule.value.type, "log")
    }
  }
}