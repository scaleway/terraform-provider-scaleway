data "scaleway_redis_cluster_versions" "redis" {
  zone               = "fr-par-1"
  include_disabled   = false
  include_beta       = false
  include_deprecated = false
}

locals {
  selected_version = data.scaleway_redis_cluster_versions.redis.versions[0].version
}
