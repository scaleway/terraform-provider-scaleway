### List default Scaleway sources

data "scaleway_cockpit_sources" "default" {
  project_id = "11111111-1111-1111-1111-111111111111"
  origin     = "scaleway"
}
