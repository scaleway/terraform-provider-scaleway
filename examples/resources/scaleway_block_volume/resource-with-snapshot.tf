### With snapshot

resource "scaleway_block_volume" "base" {
  name       = "block-volume-base"
  iops       = 5000
  size_in_gb = 20
}

resource "scaleway_block_snapshot" "main" {
  name      = "block-volume-from-snapshot"
  volume_id = scaleway_block_volume.base.id
}

resource "scaleway_block_volume" "main" {
  name        = "block-volume-from-snapshot"
  iops        = 5000
  snapshot_id = scaleway_block_snapshot.main.id
}
