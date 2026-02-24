resource "scaleway_opensearch_deployment" "prod" {
  name        = "logs-prod-cluster"
  version     = "2.0"
  node_amount = 3 # High availability with 3 nodes
  node_type   = "SEARCHDB-DEDICATED-2C-8G"
  password    = var.opensearch_password
  tags        = ["production", "logs"]

  volume {
    type       = "sbs_15k" # High IOPS for production
    size_in_gb = 100       # 100 GB
  }
}

output "opensearch_url" {
  value     = scaleway_opensearch_deployment.prod.endpoints[0].services[0].url
  sensitive = false
}
