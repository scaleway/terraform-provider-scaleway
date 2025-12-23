# Retrieve audit trail for a specific resource
data "scaleway_audit_trail_event" "find_by_resource_id" {
resource_id = "11111111-1111-1111-1111-111111111111"
}
