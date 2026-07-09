### Example with Multiple Snapshots

resource "scaleway_rdb_snapshot" "snapshot_a" {
  name        = "snapshot_a"
  instance_id = scaleway_rdb_instance.main.id
  depends_on  = [scaleway_rdb_instance.main]
}

resource "scaleway_rdb_snapshot" "snapshot_b" {
  name        = "snapshot_b"
  instance_id = scaleway_rdb_instance.main.id
  expires_at  = "2025-02-07T00:00:00Z"
  depends_on  = [scaleway_rdb_instance.main]
}
