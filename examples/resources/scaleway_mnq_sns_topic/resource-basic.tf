### Basic

resource "scaleway_mnq_sns" "main" {}

resource "scaleway_mnq_sns_credentials" "main" {
  project_id = scaleway_mnq_sns.main.project_id
  permissions {
    can_manage = true
  }
}

resource "scaleway_mnq_sns_topic" "topic" {
  project_id = scaleway_mnq_sns.main.project_id
  name       = "my-topic"
  access_key = scaleway_mnq_sns_credentials.main.access_key
  secret_key = scaleway_mnq_sns_credentials.main.secret_key
}
