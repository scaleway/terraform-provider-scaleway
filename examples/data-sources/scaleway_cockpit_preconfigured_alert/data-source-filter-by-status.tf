### Filter by status

data "scaleway_cockpit_preconfigured_alert" "enabled" {
  project_id  = scaleway_account_project.project.id
  rule_status = "enabled"
}

data "scaleway_cockpit_preconfigured_alert" "disabled" {
  project_id  = scaleway_account_project.project.id
  rule_status = "disabled"
}
