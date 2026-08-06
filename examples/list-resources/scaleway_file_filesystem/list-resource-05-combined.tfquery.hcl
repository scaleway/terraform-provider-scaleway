# List file systems with multiple filters combined
list "scaleway_file_filesystem" "combined" {
  provider = scaleway

  config {
    regios      = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    tags        = ["foobar", "barfoo"]
    name        = "db-volume"
  }
}
