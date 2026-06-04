resource "scaleway_opensearch_deployment" "main" {
  name        = "my-opensearch-cluster"
  version     = "2.0"
  node_amount = 1
  node_type   = "SEARCHDB-SHARED-2C-8G"
  password    = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}
