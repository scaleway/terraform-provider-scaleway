# Retrieve an Edge Services backend stage by its ID
data "scaleway_edge_services_backend_stage" "by_id" {
  backend_stage_id = "11111111-1111-1111-1111-111111111111"
}
