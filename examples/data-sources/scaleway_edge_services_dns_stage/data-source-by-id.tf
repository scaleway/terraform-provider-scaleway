# Retrieve an Edge Services DNS stage by its ID
data "scaleway_edge_services_dns_stage" "by_id" {
  dns_stage_id = "11111111-1111-1111-1111-111111111111"
}
