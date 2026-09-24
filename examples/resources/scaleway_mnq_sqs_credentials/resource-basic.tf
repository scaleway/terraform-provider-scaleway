### Basic

resource "scaleway_mnq_sqs" "main" {}

resource "scaleway_mnq_sqs_credentials" "main" {
  project_id = scaleway_mnq_sqs.main.project_id
  name       = "sqs-credentials"

  permissions {
    can_manage  = false
    can_receive = true
    can_publish = false
  }
}
