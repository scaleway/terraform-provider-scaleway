# List file systems filtered by name prefix
list "scaleway_file_filesystem" "by_name" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    name    = "my-fs"
  }
}
