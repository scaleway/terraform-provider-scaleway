resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-nats-trigger"
  destination_config {
    http_path   = "/ping"
    http_method = "get"
  }
  nats {
    subject                  = "TestSubject"
    server_urls              = [scaleway_mnq_nats_account.main.endpoint]
    credentials_file_content = scaleway_mnq_nats_credentials.main.file
    # If region is different
    region = scaleway_mnq_nats_account.main.region
  }
}
