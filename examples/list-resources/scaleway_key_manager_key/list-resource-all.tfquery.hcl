# List all keys across all regions and projects
list "scaleway_key_manager_key" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
