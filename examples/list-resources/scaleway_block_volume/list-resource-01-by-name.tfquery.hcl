# List block volumes filtered by name prefix
list "scaleway_block_volume" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-volume"
  }
}
