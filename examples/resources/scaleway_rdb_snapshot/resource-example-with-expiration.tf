### Example with Expiration

resource "scaleway_rdb_snapshot" "snapshot_with_expiration" {
  name        = "snapshot-with-expiration"
  instance_id = scaleway_rdb_instance.main.id
  expires_at  = "2025-01-31T00:00:00Z"
}
