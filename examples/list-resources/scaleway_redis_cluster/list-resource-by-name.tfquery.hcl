# List Redis clusters filtered by name prefix
list "scaleway_redis_cluster" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-redis"
  }
}
