# List block volumes with multiple filters combined
list "scaleway_block_volume" "combined" {
  provider = scaleway

  config {
    zones       = ["fr-par-1", "nl-ams-1"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    tags        = ["foobar", "barfoo"]
    name        = "db-volume"
  }
}
