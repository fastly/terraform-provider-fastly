# IMPORTANT: mailing list integrations require confirmation.
# To send a confirmation email and verify integration status,
# after applying changes using Terraform, please visit
# https://manage.fastly.com/observability/alerts/integrations
resource "fastly_integration" "mailinglist_example" {
  name = "my mailing list integration"
  description = "example mailing list integration"
  type = "mailinglist"

  config = {
    # mailing list address
    address = "incoming-hook@my.domain.com"
  }
}

resource "fastly_integration" "microsoftteams_example" {
  name = "my Microsoft Teams integration"
  description = "example Microsoft Teams integration"
  type = "microsoftteams"

  config = {
    # Microsoft Teams webhook URL
    webhook = "https://m365x012345.webhook.office.com"
  }
}

resource "fastly_integration" "newrelic_example" {
  name = "my New Relic integration"
  description = "example New Relic integration"
  type = "newrelic"

  config = {
    # New Relic account ID and license key
    account = "XXXXXXX"
    key = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  }
}

resource "fastly_integration" "pagerduty_example" {
  name = "my PagerDuty integration"
  description = "example PagerDuty integration"
  type = "pagerduty"

  config = {
    # PagerDuty integration key
    key = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  }
}

resource "fastly_integration" "slack_example" {
  name = "my Slack integration"
  description = "example Slack integration"
  type = "slack"

  config = {
    # Slack webhook URL
    webhook = "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
  }
}

resource "fastly_integration" "webhook_example" {
  name = "my webhook integration"
  description = "example webhook integration"
  type = "webhook"

  config = {
    # webhook URL
    webhook = "https://my.domain.com/webhook"
  }
}
