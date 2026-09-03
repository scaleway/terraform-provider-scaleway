### Basic

// For default project
data "scaleway_mnq_sns" "main" {}

// For specific project
data "scaleway_mnq_sns" "for_project" {
  project_id = scaleway_account_project.main.id
}
