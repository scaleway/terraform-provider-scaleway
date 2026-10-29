# Retrieve an Edge Services DNS stage by pipeline ID and FQDN
data "scaleway_edge_services_dns_stage" "by_fqdn" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  fqdn        = "cdn.example.com"
}
