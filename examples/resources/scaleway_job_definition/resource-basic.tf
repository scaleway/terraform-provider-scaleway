### Basic

resource "scaleway_job_definition" "main" {
  name         = "testjob"
  cpu_limit    = 140
  memory_limit = 256
  image_uri    = "docker.io/alpine:latest"
  command      = "ls"
  timeout      = "10m"

  env = {
    foo : "bar"
  }

  cron {
    schedule = "5 4 1 * *" # cron at 04:05 on day-of-month 1
    timezone = "Europe/Paris"
  }
}
