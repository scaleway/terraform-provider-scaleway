### With Dead Letter Queue

resource "scaleway_mnq_sqs" "main" {}

resource "scaleway_mnq_sqs_credentials" "main" {
  project_id = scaleway_mnq_sqs.main.project_id
  name       = "sqs-credentials"

  permissions {
    can_manage  = true
    can_receive = false
    can_publish = false
  }
}

resource "scaleway_mnq_sqs_queue" "dead_letter" {
  project_id   = scaleway_mnq_sqs.main.project_id
  name         = "dead-letter-queue"
  sqs_endpoint = scaleway_mnq_sqs.main.endpoint
  access_key   = scaleway_mnq_sqs_credentials.main.access_key
  secret_key   = scaleway_mnq_sqs_credentials.main.secret_key
}

resource "scaleway_mnq_sqs_queue" "main" {
  project_id   = scaleway_mnq_sqs.main.project_id
  name         = "my-queue"
  sqs_endpoint = scaleway_mnq_sqs.main.endpoint
  access_key   = scaleway_mnq_sqs_credentials.main.access_key
  secret_key   = scaleway_mnq_sqs_credentials.main.secret_key

  dead_letter_queue {
    id                = scaleway_mnq_sqs_queue.dead_letter.id
    max_receive_count = 3
  }
}
