# List file systems filtered by tag
list "scaleway_file_filesystem" "by_tag" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    tags    = ["bar"]
  }
}
