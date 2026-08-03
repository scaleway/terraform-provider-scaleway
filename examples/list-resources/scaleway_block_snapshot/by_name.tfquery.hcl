list "scaleway_block_snapshot" "by_name" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
    volume_ids  = ["*"]
    name        = "test-snapshot-list-1"
  }
}
