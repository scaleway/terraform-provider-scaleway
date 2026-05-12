resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-cron-trigger"
  destination_config {
    http_path   = "/patch/here"
    http_method = "patch"
  }
  cron {
    schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
    timezone = "Europe/Paris"
    body     = "{\"message\": \"This is the content to send to the container.\"}"
    headers = {
      Content-Length = 45
      Content-Type   = "application/json"
    }
  }
}
