# List Redis clusters in a specific zone for a specific project
list "scaleway_redis_cluster" "zone" {
  provider = scaleway

  config {
    zones       = ["fr-par-2"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
