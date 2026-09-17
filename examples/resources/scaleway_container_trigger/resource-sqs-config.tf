resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-sqs-trigger"
  destination_config {
    http_path   = "/"
    http_method = "get"
  }
  sqs {
    endpoint   = scaleway_mnq_sqs_queue.main.sqs_endpoint
    queue_url  = scaleway_mnq_sqs_queue.main.url
    access_key = scaleway_mnq_sqs_credentials.main.access_key
    secret_key = scaleway_mnq_sqs_credentials.main.secret_key
    # If region is different
    region = scaleway_mnq_sqs.main.region
  }
}
