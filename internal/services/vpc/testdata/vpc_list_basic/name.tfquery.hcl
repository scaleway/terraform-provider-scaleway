list "scaleway_vpc" "by_name" {
  provider = scaleway

  config {
    project_id = scaleway_account_project.main.id
    region = "all"
    name = "test-vpc"
  }
}
