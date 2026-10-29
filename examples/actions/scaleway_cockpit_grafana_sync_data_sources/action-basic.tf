### Synchronize Cockpit data sources with Grafana after source creation

resource "scaleway_cockpit_source" "metrics" {
  project_id     = var.project_id
  name           = "prod-metrics"
  type           = "metrics"
  retention_days = 31

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.scaleway_cockpit_grafana_sync_data_sources.sync]
    }
  }
}

action "scaleway_cockpit_grafana_sync_data_sources" "sync" {
  config {
    project_id = var.project_id
  }
}
