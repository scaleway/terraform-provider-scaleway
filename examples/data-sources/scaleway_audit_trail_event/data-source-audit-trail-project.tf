# Retrieve audit trail events on a specific project
data "scaleway_audit_trail_event" "find_by_project" {
  project_id = "11111111-1111-1111-1111-111111111111"
}
