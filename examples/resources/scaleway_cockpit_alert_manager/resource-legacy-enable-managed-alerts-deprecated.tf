### Legacy: Enable managed alerts (Deprecated)

resource "scaleway_account_project" "project" {
  name = "tf_test_project"
}

resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id            = scaleway_account_project.project.id
  enable_managed_alerts = true

  contact_points {
    email = "alert@example.com"
  }
}
