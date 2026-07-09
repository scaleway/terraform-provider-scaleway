list "scaleway_block_snapshot" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
    volume_ids  = ["*"]
  }
}
