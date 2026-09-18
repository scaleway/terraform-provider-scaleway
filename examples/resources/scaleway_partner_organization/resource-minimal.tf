resource "scaleway_partner_organization" "minimal" {
  email             = "contact@example.com"
  organization_name = "Minimal Organization"
  partner_id        = "11111111-1111-1111-1111-111111111111"
  owner_firstname   = "John"
  owner_lastname    = "Doe"
  customer_id       = "customer-minimal"
}
