### Basic

// For default project
data "scaleway_mnq_sqs" "main" {}

// For specific project
data "scaleway_mnq_sqs" "for_project" {
  project_id = scaleway_account_project.main.id
}
