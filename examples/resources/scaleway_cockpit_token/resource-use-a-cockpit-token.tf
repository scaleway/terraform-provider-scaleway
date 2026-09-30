### Use a Cockpit token

resource "scaleway_account_project" "project" {
  name = "my-project"
}

resource "scaleway_cockpit_token" "main" {
  project_id = scaleway_account_project.project.id
  name       = "my-awesome-token"
}
