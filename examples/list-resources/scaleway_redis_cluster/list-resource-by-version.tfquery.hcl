# List Redis clusters filtered by engine version
list "scaleway_redis_cluster" "by_version" {
  provider = scaleway

  config {
    zones       = ["*"]
    project_ids = ["*"]
    version     = "7.2"
  }
}
