# List OpenSearch deployments filtered by name prefix
list "scaleway_opensearch_deployment" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "my-opensearch"
  }
}
