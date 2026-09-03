### Use a Cockpit token

resource "scaleway_account_project" "project" {
  name = "my-project"
}

// Create a token that can read metrics and logs but not write
resource "scaleway_cockpit_token" "main" {
  project_id = scaleway_account_project.project.id

  name = "my-awesome-token"
  scopes {
    query_metrics = true
    write_metrics = false

    query_logs = true
    write_logs = false
  }
}
