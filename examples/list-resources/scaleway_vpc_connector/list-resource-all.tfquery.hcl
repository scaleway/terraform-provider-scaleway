# List VPC connectors across all regions and all projects
list "scaleway_vpc_connector" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
