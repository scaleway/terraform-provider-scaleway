list "scaleway_block_snapshot" "by_volume" {
  provider = scaleway

  config {
    zones       = [scaleway_block_volume.vol1.zone]
    project_ids = [scaleway_block_volume.vol1.project_id]
    volume_ids  = [scaleway_block_volume.vol1.id]
  }
}
