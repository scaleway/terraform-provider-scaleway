### Use the Grafana Terraform provider (Deprecated)

// Old approach (deprecated) - Using scaleway_cockpit_grafana_user
// resource "scaleway_cockpit_grafana_user" "main" {
//   project_id = scaleway_cockpit.main.project_id
//   login      = "example"
//   role       = "editor"
// }
//
// provider "grafana" {
//   url  = scaleway_cockpit.main.endpoints.0.grafana_url
//   auth = "${scaleway_cockpit_grafana_user.main.login}:${scaleway_cockpit_grafana_user.main.password}"
// }

// New approach - Use scaleway_cockpit_grafana data source with IAM auth
data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_cockpit.main.project_id
}

// Note: Grafana provider with IAM auth requires proper token setup
// See Grafana provider documentation for IAM authentication
