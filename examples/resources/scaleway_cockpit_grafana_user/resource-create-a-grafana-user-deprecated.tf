### Create a Grafana user (Deprecated)

resource "scaleway_account_project" "project" {
  name = "test project grafana user"
}

resource "scaleway_cockpit_grafana_user" "main" {
  project_id = scaleway_account_project.project.id
  login      = "my-awesome-user"
  role       = "editor"
}
