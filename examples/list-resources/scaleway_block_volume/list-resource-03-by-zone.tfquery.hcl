# List block volumes in a specific zone
list "scaleway_block_volume" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
