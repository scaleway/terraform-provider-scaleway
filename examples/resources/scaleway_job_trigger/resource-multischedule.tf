resource scaleway_job_definition main {
  cpu_limit = 120
  memory_limit = 256
  local_storage_capacity = 5120
  image_uri = "docker.io/alpine:latest"
}

resource "scaleway_job_trigger" "main" {
  name              = "job-trigger-main"
  job_definition_id = scaleway_job_definition.main.id

  cron_config = {
    schedule        = "0 8 * * *" # cron at 8AM every day
    timezone        = "Europe/Paris"
    startup_command = ["echo"]
    args            = ["starting", "daily", "job"]
  }
}

resource "scaleway_job_trigger" "secondary" {
  name              = "job-trigger-secondary"
  job_definition_id = scaleway_job_definition.main.id

  cron_config = {
    schedule        = "0 17 * * *" # cron at 5PM every day
    timezone        = "Europe/Paris"
    startup_command = ["echo"]
    args            = ["job", "done"]
  }
}
