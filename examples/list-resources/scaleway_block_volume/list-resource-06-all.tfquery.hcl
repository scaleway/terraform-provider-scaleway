# List block volumes across all zones and all projects
list "scaleway_block_volume" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
