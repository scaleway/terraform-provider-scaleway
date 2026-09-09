resource "scaleway_billing_budget" "main" {
  organization_id   = "11111111-1111-1111-1111-111111111111"
  consumption_limit = 10000
  enabled           = true
}

resource "scaleway_billing_budget_alert" "main" {
  budget_id = scaleway_billing_budget.main.id
  threshold = 80
}
