### Datadog destination

data "scaleway_account_project" "project" {
  name = "default"
}

data "scaleway_cockpit_sources" "scaleway_metrics" {
  project_id = data.scaleway_account_project.project.id
  origin     = "scaleway"
  type       = "metrics"
}

resource "scaleway_cockpit_exporter" "main" {
  project_id        = data.scaleway_account_project.project.id
  datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
  name              = "my-datadog-exporter"
  exported_products = ["all"]

  datadog_destination {
    api_key  = var.datadog_api_key
    endpoint = "https://api.datadoghq.com"
  }
}
