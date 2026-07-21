# List block volumes filtered by tag
list "scaleway_block_volume" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["bar"]
  }
}
