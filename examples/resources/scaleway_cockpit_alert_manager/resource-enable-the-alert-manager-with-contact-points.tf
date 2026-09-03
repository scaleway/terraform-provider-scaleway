### Enable the alert manager with contact points

resource "scaleway_account_project" "project" {
  name = "tf_test_project"
}

resource "scaleway_cockpit_alert_manager" "alert_manager" {
  project_id = scaleway_account_project.project.id

  contact_points {
    email = "alert1@example.com"
  }

  contact_points {
    email = "alert2@example.com"
  }
}
