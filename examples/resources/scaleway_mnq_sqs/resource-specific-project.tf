### Specific Project

data "scaleway_account_project" "project" {
  name = "default"
}

resource "scaleway_mnq_sqs" "for_project" {
  project_id = data.scaleway_account_project.project.id
}
