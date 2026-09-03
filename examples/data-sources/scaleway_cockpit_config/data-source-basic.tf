data "scaleway_cockpit_config" "main" {
  region = "fr-par"
}

resource "scaleway_cockpit_source" "metrics" {
  name           = "my-metrics"
  type           = "metrics"
  retention_days = data.scaleway_cockpit_config.main.custom_metrics_retention[0].default_days
}

output "custom_metrics_retention_bounds" {
  value = data.scaleway_cockpit_config.main.custom_metrics_retention[0]
}
