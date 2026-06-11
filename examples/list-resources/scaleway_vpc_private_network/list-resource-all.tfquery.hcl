# List Private Networks across all regions and all projects
list "scaleway_vpc_private_network" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
