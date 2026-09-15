### Specific Project

data "scaleway_account_project" "project" {
  name = "default"
}

// For specific Project in default region
resource "scaleway_mnq_sns" "for_project" {
  project_id = data.scaleway_account_project.project.id
}
