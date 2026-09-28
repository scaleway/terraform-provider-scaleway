resource "scaleway_partner_organization" "main" {
  email             = "contact@example.com"
  organization_name = "My Organization"
  partner_id        = "11111111-1111-1111-1111-111111111111"
  owner_firstname   = "John"
  owner_lastname    = "Doe"
  customer_id       = "customer-123"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_partner_organization_request_admin_role.main]
    }
  }
}

action "scaleway_partner_organization_request_admin_role" "main" {
  config {
    organization_id = scaleway_partner_organization.main.id
    username        = "partner_user"
    email           = "partner@example.com"
    password        = "SecurePassword123!"
  }
}
