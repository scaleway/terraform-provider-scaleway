list "scaleway_vpc" "fr-par" {
  provider = scaleway

  config {
    project_id = scaleway_account_project.main.id
    region = "fr-par"
  }
}
