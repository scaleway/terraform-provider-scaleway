### Create a snapshot of a Block Storage volume

resource "scaleway_block_volume" "block_volume" {
  iops       = 5000
  name       = "some-volume-name"
  size_in_gb = 20
}

resource "scaleway_block_snapshot" "block_snapshot" {
  name      = "some-snapshot-name"
  volume_id = scaleway_block_volume.block_volume.id
}
