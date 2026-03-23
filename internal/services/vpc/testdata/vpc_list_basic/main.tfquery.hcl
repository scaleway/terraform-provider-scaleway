list "scaleway_vpc" "all" {
  provider = scaleway

  config {
    region = "all"
    project_id = scaleway_account_project.main.id
  }
}

list "scaleway_vpc" "fr-par" {
  provider = scaleway

  config {
    project_id = scaleway_account_project.main.id
    region = "fr-par"
    tags = ["environment=production"]
  }
}

list "scaleway_vpc" "by_name" {
  provider = scaleway

  config {
    project_id = scaleway_account_project.main.id
    region = "all"
    name = "test-vpc"
  }
}
