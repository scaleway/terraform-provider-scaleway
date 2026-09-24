### Trigger a test alert on a Cockpit

resource "scaleway_cockpit" "main" {
  project_id = var.project_id
  name       = "my-cockpit"
}

action "scaleway_cockpit_trigger_test_alert" "test_alert" {
  config {
    cockpit_id = scaleway_cockpit.main.id
  }
}
