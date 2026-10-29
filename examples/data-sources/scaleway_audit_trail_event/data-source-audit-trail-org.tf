# Retrieve audit trail events on a specific organization
data "scaleway_audit_trail_event" "find_by_org" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}
