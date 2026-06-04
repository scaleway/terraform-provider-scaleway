resource "scaleway_opensearch_deployment" "analytics" {
  name        = "analytics-cluster"
  version     = "2.0"
  node_amount = 1
  node_type   = "SEARCHDB-SHARED-4C-16G"
  password    = var.opensearch_password
  tags        = ["analytics", "dev", "team-data"]

  volume {
    type       = "sbs_5k"
    size_in_gb = 10
  }
}
