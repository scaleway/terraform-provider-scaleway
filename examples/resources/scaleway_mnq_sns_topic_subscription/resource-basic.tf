### Basic

// For default project in default region
resource "scaleway_mnq_sns" "main" {}

resource "scaleway_mnq_sns_credentials" "main" {
  project_id = scaleway_mnq_sns.main.project_id
  permissions {
    can_manage  = true
    can_publish = true
    can_receive = true
  }
}

resource "scaleway_mnq_sns_topic" "topic" {
  project_id = scaleway_mnq_sns.main.project_id
  name       = "my-topic"
  access_key = scaleway_mnq_sns_credentials.main.access_key
  secret_key = scaleway_mnq_sns_credentials.main.secret_key
}

resource "scaleway_mnq_sns_topic_subscription" "main" {
  project_id = scaleway_mnq_sns.main.project_id
  access_key = scaleway_mnq_sns_credentials.main.access_key
  secret_key = scaleway_mnq_sns_credentials.main.secret_key
  topic_id   = scaleway_mnq_sns_topic.topic.id
  protocol   = "http"
  endpoint   = "http://example.com"
}
