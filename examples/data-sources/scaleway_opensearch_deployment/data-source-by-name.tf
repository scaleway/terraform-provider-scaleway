# Get info by name
data "scaleway_opensearch_deployment" "by_name" {
  name = "my-opensearch-cluster"
}
