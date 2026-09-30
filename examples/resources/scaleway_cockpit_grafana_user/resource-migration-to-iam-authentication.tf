### Migration to IAM Authentication

# Old approach (deprecated)
# resource "scaleway_cockpit_grafana_user" "main" {
#   project_id = scaleway_account_project.project.id
#   login      = "my-awesome-user"
#   role       = "editor"
# }

# New approach - Use IAM authentication
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

output "grafana_url" {
  value       = data.scaleway_cockpit_grafana.main.grafana_url
  description = "Access Grafana using your Scaleway IAM credentials"
}
