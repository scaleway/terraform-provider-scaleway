# List OpenSearch deployments filtered by engine version
list "scaleway_opensearch_deployment" "by_version" {
  provider = scaleway

  config {
    regions     = ["*"]
    project_ids = ["*"]
    version     = "2.15"
  }
}
