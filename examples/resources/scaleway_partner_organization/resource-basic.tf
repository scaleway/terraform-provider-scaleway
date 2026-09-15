resource "scaleway_partner_organization" "main" {
  email             = "contact@example.com"
  organization_name = "My Organization"
  partner_id        = "11111111-1111-1111-1111-111111111111"
  owner_firstname   = "John"
  owner_lastname    = "Doe"
  customer_id       = "customer-123"
  phone_number      = "+33123456789"
}
