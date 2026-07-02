# List OpenSearch deployments filtered by tag
list "scaleway_opensearch_deployment" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
