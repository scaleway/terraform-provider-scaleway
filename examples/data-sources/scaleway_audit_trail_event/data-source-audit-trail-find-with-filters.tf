# Retrieve audit trail events with various filtering
data "scaleway_audit_trail_event" "find_with_filters" {
  region          = "fr-par"
  service_name    = "instance"
  method_name     = "CreateServer"
  principal_id    = "11111111-1111-1111-1111-111111111111"
  source_ip       = "192.0.2.1"
  status          = 200
  recorded_after  = "2025-10-01T00:00:00Z"
  recorded_before = "2025-12-31T23:59:59Z"
  order_by        = "recorded_at_desc"
}
