resource "fastly_ngwaf_account_rule" "example" {
  applies_to       = ["*"]
  type             = "request"
  description      = "example"
  enabled          = true
  request_logging  = "sampled"
  group_operator   = "all"

  action {
    type = "block"
  }

  condition {
    field    = "ip"
    operator = "equals"
    value    = "1.2.3.4"
  }

  group_condition {
    group_operator = "all"

    condition {
      field    = "method"
      operator = "equals"
      value    = "POST"
    }
  }
}
