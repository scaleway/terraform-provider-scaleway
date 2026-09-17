# List all IPAM IPs across all regions and all projects
list "scaleway_ipam_ip" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
