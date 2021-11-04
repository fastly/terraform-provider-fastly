data "fastly_waf_rules" "tag" {
  tags = ["language-html", "language-jsp"]
}