### Basic

resource "scaleway_mnq_sns" "main" {}

resource "scaleway_mnq_sns_credentials" "main" {
  project_id = scaleway_mnq_sns.main.project_id
  name       = "sns-credentials"

  permissions {
    can_manage  = false
    can_receive = true
    can_publish = false
  }
}
