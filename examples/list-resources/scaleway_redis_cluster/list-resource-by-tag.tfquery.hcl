# List Redis clusters filtered by tag
list "scaleway_redis_cluster" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["production"]
  }
}
