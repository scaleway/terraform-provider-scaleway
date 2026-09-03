### OTLP destination

data "scaleway_account_project" "project" {
  name = "default"
}

data "scaleway_cockpit_sources" "scaleway_metrics" {
  project_id = data.scaleway_account_project.project.id
  origin     = "scaleway"
  type       = "metrics"
}

resource "scaleway_cockpit_source" "otlp_target" {
  project_id     = data.scaleway_account_project.project.id
  name           = "otlp-target"
  type           = "metrics"
  retention_days = 31
}

resource "scaleway_cockpit_exporter" "main" {
  project_id        = data.scaleway_account_project.project.id
  datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
  name              = "my-otlp-exporter"
  exported_products = ["lb", "object-storage", "rdb"]

  otlp_destination {
    endpoint = scaleway_cockpit_source.otlp_target.push_url
  }
}
