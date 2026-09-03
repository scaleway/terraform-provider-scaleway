#### From snapshot

data "scaleway_block_snapshot" "snapshot" {
  name = "my_snapshot"
}

resource "scaleway_block_volume" "from_snapshot" {
  snapshot_id = data.scaleway_block_snapshot.snapshot.id
  iops        = 5000
}

resource "scaleway_instance_server" "from_snapshot" {
  type = "PRO2-XXS"
  root_volume {
    volume_id   = scaleway_block_volume.from_snapshot.id
    volume_type = "sbs_volume"
  }
}
