list "scaleway_block_snapshot" "by_zone" {
  provider = scaleway

  config {
    volume_ids  = ["*"]
    project_ids = ["*"]
    zones       = ["pl-waw-2"]
  }
}
