### Using with default project

# Uses the default project from provider configuration
data "scaleway_cockpit_grafana" "main" {}

output "grafana_url" {
  value = data.scaleway_cockpit_grafana.main.grafana_url
}
