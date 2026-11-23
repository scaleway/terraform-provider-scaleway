resource "scaleway_job_definition" "main" {
  cpu_limit = 140
  memory_limit = 256
  local_storage_capacity = 5120
  image_uri    = "docker.io/alpine:latest"
}

resource "scaleway_job_trigger" "main" {
  name              = "job-trigger"
  job_definition_id = scaleway_job_definition.main.id

  cron_config = {
    schedule = "5 4 1 * *" # cron at 04:05 on day-of-month 1
    timezone = "Europe/Paris"
  }
}
