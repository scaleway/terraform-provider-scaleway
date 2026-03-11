# Get info by deployment ID
data "scaleway_opensearch_deployment" "by_id" {
  deployment_id = "11111111-1111-1111-1111-111111111111"
}
