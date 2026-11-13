# List Public Gateway IPs across all projects in a zone
list "scaleway_vpc_public_gateway_ip" "by_project" {
  provider = scaleway

  config {
    zones       = ["fr-par-1"]
    project_ids = ["*"]
  }
}
