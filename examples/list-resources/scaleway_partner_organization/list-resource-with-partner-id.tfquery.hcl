# List organizations managed by the partner, specifying the partner ID explicitly
list "scaleway_partner_organization" "by_partner" {
  provider = scaleway

  config {
    partner_id  = "11111111-1111-1111-1111-111111111111"
    customer_id = "customer-123"
  }
}
