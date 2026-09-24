### Complete example with Cockpit setup

resource "scaleway_account_project" "project" {
  name = "my-observability-project"
}

resource "scaleway_cockpit" "main" {
  project_id = scaleway_account_project.project.id
}

data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_cockpit.main.project_id
}

output "grafana_connection_info" {
  value = {
    url        = data.scaleway_cockpit_grafana.main.grafana_url
    project_id = data.scaleway_cockpit_grafana.main.project_id
  }
  description = "Use your Scaleway IAM credentials to authenticate"
}
