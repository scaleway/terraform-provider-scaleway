# Retrieve an Edge Services TLS stage by its ID
data "scaleway_edge_services_tls_stage" "by_id" {
  tls_stage_id = "11111111-1111-1111-1111-111111111111"
}
