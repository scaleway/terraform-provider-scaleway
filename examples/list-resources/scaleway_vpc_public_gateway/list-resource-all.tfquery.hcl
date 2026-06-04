# List Public Gateways across all zones and all projects
list "scaleway_vpc_public_gateway" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
