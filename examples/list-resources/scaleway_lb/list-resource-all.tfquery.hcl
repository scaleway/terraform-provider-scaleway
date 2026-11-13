# List Load Balancers across all zones and all projects
list "scaleway_lb" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
