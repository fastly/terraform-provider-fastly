data "fastly_waf_rules" "owasp_with_exclusions" {
  publishers              = ["owasp"]
  exclude_modsec_rule_ids = [1010090]
}