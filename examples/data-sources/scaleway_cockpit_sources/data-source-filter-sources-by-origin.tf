### Filter sources by origin

data "scaleway_cockpit_sources" "custom" {
  project_id = "11111111-1111-1111-1111-111111111111"
  origin     = "custom"
}
