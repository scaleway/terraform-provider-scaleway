### Export an instance snapshot to Object Storage

resource "scaleway_instance_snapshot" "main" {
  name      = "my-instance-snapshot"
  volume_id = scaleway_instance_volume.main.id
  bucket    = scaleway_object_bucket.main.name
  key       = "snapshots/my-snapshot"
}

resource "scaleway_instance_volume" "main" {
  name       = "my-volume"
  type       = "b_ssd"
  size_in_gb = 20
}

resource "scaleway_object_bucket" "main" {
  name = "my-export-bucket"
}

action "scaleway_instance_export_snapshot" "export" {
  config {
    snapshot_id = scaleway_instance_snapshot.main.id
    wait        = true
  }
}
