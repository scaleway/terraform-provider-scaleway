# List MongoDB instances across all regions and all projects
list "scaleway_mongodb_instance" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
