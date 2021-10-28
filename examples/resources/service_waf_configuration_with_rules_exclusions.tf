resource "fastly_service_v1" "demo" {
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

resource "fastly_service_waf_configuration" "waf" {
  waf_id                         = fastly_service_v1.demo.waf[0].waf_id
  http_violation_score_threshold = 100

  rule {
    modsec_rule_id = 2029718
    revision       = 1
    status         = "log"
  }

  rule_exclusion {
    name            = "index page"
    exclusion_type  = "rule"
    condition       = "req.url.basename == \"index.html\""
    modsec_rule_ids = [2029718]
  }
}