# Retrieve an Edge Services WAF stage by its ID
data "scaleway_edge_services_waf_stage" "by_id" {
  waf_stage_id = "11111111-1111-1111-1111-111111111111"
}
