### Retrieve a data source by filters

data "scaleway_cockpit_source" "filtered" {
  project_id = "11111111-1111-1111-1111-111111111111"
  region     = "fr-par"
  name       = "my-data-source"
}
