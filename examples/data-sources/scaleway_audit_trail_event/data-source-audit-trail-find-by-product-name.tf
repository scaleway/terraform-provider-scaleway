# Retrieve audit trail for a specific Scaleway product
data "scaleway_audit_trail_event" "find_by_product_name" {
  product_name = "secret-manager"
}
