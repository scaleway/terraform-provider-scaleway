# Retrieve audit trail events for a specific type of resource
data "scaleway_audit_trail_event" "find_by_resource_type" {
  resource_type = "instance_server"
}
