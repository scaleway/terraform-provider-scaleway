### MongoDB instance restored from Snapshot

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5
}

resource "scaleway_mongodb_snapshot" "main_snapshot" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "my-mongodb-snapshot"
}

resource "scaleway_mongodb_instance" "restored_instance" {
  snapshot_id = scaleway_mongodb_snapshot.main_snapshot.id
  name        = "restored-mongodb-from-snapshot"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
}
