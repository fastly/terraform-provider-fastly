resource "fastly_ngwaf_workspace" "demo" {
  name                         = "Demo"
  description                  = "Testing"
  mode                         = "block"

  attack_signal_thresholds {
    one_minute  = 100
    ten_minutes = 500
    one_hour    = 1000
    immediate   = true
  }
}
