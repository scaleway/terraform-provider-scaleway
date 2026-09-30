list "scaleway_block_snapshot" "by_tag" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
    volume_ids  = ["*"]
    tags        = ["test-tag"]
  }
}
