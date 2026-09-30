### Basic

resource "scaleway_edge_services_dns_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  fqdns       = ["subdomain.example.com"]
}
