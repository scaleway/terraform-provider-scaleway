### Filter sources by name

data "scaleway_cockpit_sources" "my_sources" {
  project_id = "11111111-1111-1111-1111-111111111111"
  name       = "my-data-source"
}
