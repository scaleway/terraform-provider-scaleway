### Filter sources by type

data "scaleway_cockpit_sources" "metrics" {
  project_id = "11111111-1111-1111-1111-111111111111"
  type       = "metrics"
}
