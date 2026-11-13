# List VPCs across all regions and all projects
list "scaleway_vpc" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
