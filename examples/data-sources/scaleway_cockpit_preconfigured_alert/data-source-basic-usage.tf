### Basic usage

data "scaleway_cockpit_preconfigured_alert" "main" {
  project_id = scaleway_account_project.project.id
}

output "available_alerts" {
  value = data.scaleway_cockpit_preconfigured_alert.main.alerts
}
