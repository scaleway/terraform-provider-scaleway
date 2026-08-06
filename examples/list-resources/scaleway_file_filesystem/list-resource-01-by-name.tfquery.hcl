# List file systems filtered by name prefix
list "scaleway_block_volume" "by_name" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    name    = "my-fs"
  }
}
