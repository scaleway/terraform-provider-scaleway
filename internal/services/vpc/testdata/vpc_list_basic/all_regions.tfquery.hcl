list "scaleway_vpc" "all" {
  provider = scaleway

  config {
    region = "all"
    project_id = scaleway_account_project.main.id
  }
}
