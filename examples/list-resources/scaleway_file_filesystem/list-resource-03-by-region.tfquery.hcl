# List file systems in a specific region
list "scaleway_file_filesystem" "by_region" {
  provider = scaleway

  config {
    regions = ["fr-par"] # Only one supported for now
  }
}
