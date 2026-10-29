# List organizations managed by the partner filtered by customer ID
list "scaleway_partner_organization" "by_customer" {
  provider = scaleway

  config {
    customer_id = "customer-123"
  }
}
