# List Redis clusters across all zones and all projects
list "scaleway_redis_cluster" "all" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
  }
}
