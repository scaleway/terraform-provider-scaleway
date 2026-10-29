# List OpenSearch deployments across all regions and all projects
list "scaleway_opensearch_deployment" "all" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
  }
}
