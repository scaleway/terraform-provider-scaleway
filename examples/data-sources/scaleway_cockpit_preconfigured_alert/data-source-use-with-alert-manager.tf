### Use with Alert Manager

resource "scaleway_account_project" "project" {
  name = "my-observability-project"
}

resource "scaleway_cockpit" "main" {
  project_id = scaleway_account_project.project.id
}

data "scaleway_cockpit_preconfigured_alert" "all" {
  project_id = scaleway_cockpit.main.project_id
}

resource "scaleway_cockpit_alert_manager" "main" {
  project_id = scaleway_cockpit.main.project_id

  # Enable specific alerts by their preconfigured_rule_id
  preconfigured_alert_ids = [
    for alert in data.scaleway_cockpit_preconfigured_alert.all.alerts :
    alert.preconfigured_rule_id
    if alert.product_name == "instance" && alert.rule_status == "disabled"
  ]

  contact_points {
    email = "alerts@example.com"
  }
}
