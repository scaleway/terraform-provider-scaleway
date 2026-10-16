### Activate Grafana for a project after creation

resource "scaleway_account_project" "project" {
  name = "observability"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_cockpit_activate_grafana.main]
    }
  }
}

action "scaleway_cockpit_activate_grafana" "main" {
  config {
    project_id = scaleway_account_project.project.id
  }
}
