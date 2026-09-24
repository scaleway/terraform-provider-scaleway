### Create a data source

resource "scaleway_account_project" "project" {
  name = "test project data source"
}

resource "scaleway_cockpit_source" "main" {
  project_id     = scaleway_account_project.project.id
  name           = "my-data-source"
  type           = "metrics"
  retention_days = 6
}
