resource "scaleway_billing_budget" "main" {
  organization_id   = "11111111-1111-1111-1111-111111111111"
  consumption_limit = 10000
  enabled           = true
}

resource "scaleway_billing_budget_alert" "main" {
  budget_id = scaleway_billing_budget.main.id
  threshold = 80
}

resource "scaleway_billing_budget_alert_notification" "email" {
  budget_alert_id = scaleway_billing_budget_alert.main.id
  email_addresses = ["alerts@example.com", "billing@example.com"]
}

resource "scaleway_billing_budget_alert_notification" "sms" {
  budget_alert_id   = scaleway_billing_budget_alert.main.id
  sms_phone_numbers = ["+33612345678"]
}

resource "scaleway_billing_budget_alert_notification" "webhook" {
  budget_alert_id = scaleway_billing_budget_alert.main.id
  webhook_urls    = ["https://example.com/webhook"]
}
