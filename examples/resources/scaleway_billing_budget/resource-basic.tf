resource "scaleway_billing_budget" "main" {
  organization_id   = "11111111-1111-1111-1111-111111111111"
  consumption_limit = 100
  enabled           = true
}
