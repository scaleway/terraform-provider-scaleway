# List all servers of a specific project in all zones
list "scaleway_instance_server" "by_project" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}