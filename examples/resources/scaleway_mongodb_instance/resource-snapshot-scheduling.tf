### MongoDB instance with Snapshot Scheduling

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-with-snapshots"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5

  # Snapshot scheduling configuration
  snapshot_schedule_frequency_hours = 24
  snapshot_schedule_retention_days  = 7
  is_snapshot_schedule_enabled      = true
}
