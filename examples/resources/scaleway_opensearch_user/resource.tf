resource "scaleway_opensearch_deployment" "main" {
  name       = "my-opensearch-cluster"
  version    = "2.0"
  node_count = 1
  node_type  = "SEARCHDB-SHARED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

resource "scaleway_opensearch_user" "app" {
  deployment_id = scaleway_opensearch_deployment.main.id
  name          = "app_user"
  password      = "ThisIsASecurePassword123!"
}
