### Export a block snapshot to Object Storage

resource "scaleway_block_snapshot" "main" {
  name       = "my-block-snapshot"
  volume_id  = scaleway_block_volume.main.id
  bucket_id  = scaleway_object_bucket.main.id
  object_key = "snapshots/my-snapshot"
}

resource "scaleway_block_volume" "main" {
  name             = "my-volume"
  size_in_gb       = 20
  performance_mode = "high"
  volume_class     = "unified"
}

resource "scaleway_object_bucket" "main" {
  name = "my-export-bucket"
}

action "scaleway_block_export_snapshot" "export" {
  config {
    snapshot_id = scaleway_block_snapshot.main.id
    wait        = true
  }
}
